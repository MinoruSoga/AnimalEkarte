package reservation

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLineSettingRepo struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	findTokenFn      func(ctx context.Context, clinicID uint64) (string, error)
}

func (m *mockLineSettingRepo) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{
		ClinicID:          clinicID,
		NotificationEmail: "notify@example.com",
	}, nil
}

func (m *mockLineSettingRepo) FindLineChannelAccessTokenCiphertext(ctx context.Context, clinicID uint64) (string, error) {
	if m.findTokenFn != nil {
		return m.findTokenFn(ctx, clinicID)
	}
	return "", nil
}

func TestReservationNotificationService_NotifyCreated(t *testing.T) {
	ctx := context.Background()

	t.Run("success notifying created", func(t *testing.T) {
		done := make(chan uint64, 1)
		repo := &mockLineSettingRepo{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				setting := &model.LineReservationSetting{
					ClinicID:          clinicID,
					NotificationEmail: "notify@example.com",
				}
				done <- clinicID
				return setting, nil
			},
		}
		cfg := &ReservationNotificationConfig{
			SMTPHost: "",
		}
		svc := NewReservationNotificationService(cfg, repo, nil, nil, nil)

		appt := &model.Reservation{
			ID:        123,
			ClinicID:  1,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Pet:       &model.Pet{Name: "ポチ"},
			Doctor:    &model.Staff{Name: "獣医師A"},
			ReservationType: &model.ReservationType{
				Name: "一般診療",
			},
			CustomerFields: []byte(`{"pets": [{"name": "ポチ", "type": "犬"}]}`),
		}
		customer := &model.LineCustomer{
			LineUserID:       "U12345",
			DisplayName:      "山田太郎",
			RealName:         "山田太郎本名",
			AdditionalFields: []byte(`{"phone": "090-0000-0000", "note": "かゆみ"}`),
		}

		svc.NotifyCreated(ctx, appt, customer)

		select {
		case gotClinicID := <-done:
			// 到達assert: 医院設定が正しく参照され、送信判定（LINE/メール）に到達したことを確認する。
			// 実ネットワーク送信（LINE Push）完了までは待たない（-race・タイムアウトの安定性のため）。
			assert.Equal(t, appt.ClinicID, gotClinicID, "notification should look up settings for the reservation's clinic")
		case <-time.After(30 * time.Second):
			t.Fatal("timed out waiting for NotifyCreated to reach clinic setting lookup")
		}
	})

	t.Run("db setting load error", func(t *testing.T) {
		var called bool
		repo := &mockLineSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				called = true
				return nil, errors.New("db load error")
			},
		}
		cfg := &ReservationNotificationConfig{}
		svc := NewReservationNotificationService(cfg, repo, nil, nil, nil)

		appt := &model.Reservation{
			ID:       123,
			ClinicID: 1,
		}

		svc.NotifyCreated(ctx, appt, nil)
		svc.Wait()

		// 到達assert: setting load 失敗時は LINE/メール送信を試みず早期returnする（送信されなかった分岐）。
		// Wait() が速やかに返ること自体が「後続のネットワーク送信を試みていない」ことの証跡。
		assert.True(t, called, "clinic setting lookup should have been attempted")
	})
}

func TestReservationNotificationService_NotifyCancelled(t *testing.T) {
	ctx := context.Background()

	t.Run("success notifying cancelled", func(t *testing.T) {
		done := make(chan uint64, 1)
		repo := &mockLineSettingRepo{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				setting := &model.LineReservationSetting{
					ClinicID:          clinicID,
					NotificationEmail: "notify@example.com",
				}
				done <- clinicID
				return setting, nil
			},
		}
		cfg := &ReservationNotificationConfig{
			SMTPHost: "",
		}
		svc := NewReservationNotificationService(cfg, repo, nil, nil, nil)

		appt := &model.Reservation{
			ID:        123,
			ClinicID:  1,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
		}
		customer := &model.LineCustomer{
			LineUserID:  "U12345",
			DisplayName: "山田太郎",
		}

		svc.NotifyCancelled(ctx, appt, customer)

		select {
		case gotClinicID := <-done:
			assert.Equal(t, appt.ClinicID, gotClinicID, "notification should look up settings for the reservation's clinic")
		case <-time.After(30 * time.Second):
			t.Fatal("timed out waiting for NotifyCancelled to reach clinic setting lookup")
		}
	})

	t.Run("db setting load error", func(t *testing.T) {
		var called bool
		repo := &mockLineSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				called = true
				return nil, errors.New("db load error")
			},
		}
		cfg := &ReservationNotificationConfig{}
		svc := NewReservationNotificationService(cfg, repo, nil, nil, nil)

		appt := &model.Reservation{
			ID:       123,
			ClinicID: 1,
		}

		svc.NotifyCancelled(ctx, appt, nil)
		svc.Wait()

		assert.True(t, called, "clinic setting lookup should have been attempted")
	})
}

// fakeLinePusher は LinePusher のテスト用実装（Push 先とテキストを記録し、設定された error を返す）。
type fakeLinePusher struct {
	err    error
	pushes []fakeLinePush
}

type fakeLinePush struct {
	to   string
	text string
}

func (f *fakeLinePusher) PushText(_ context.Context, to, text string) error {
	f.pushes = append(f.pushes, fakeLinePush{to: to, text: text})
	return f.err
}

// EMR-504 Q5(a): 予約の完了・キャンセル通知は、clinic_integrations のトークン（暗号文）を復号して
// 予約本人（line_customers）自身の LINE ユーザー ID 宛へ直接 Push する best-effort 通知である。
// トークン未設定・読取り/復号/送信の失敗は Push だけを止め、予約もクリニック宛メールも止めない。
func TestReservationNotificationService_LinePush(t *testing.T) {
	const (
		storedCiphertext = "ciphertext-aaaa"
		plainToken       = "token-bbbb"
		customerLineID   = "U-customer-cccc"
	)
	ctx := context.Background()
	tests := []struct {
		name       string
		ciphertext string
		tokenErr   error
		decryptErr error
		pushErr    error
		wantPush   bool
	}{
		{name: "pushes with the decrypted token to the reserving customer's own LINE id", ciphertext: storedCiphertext, wantPush: true},
		{name: "token absent: no push", ciphertext: ""},
		{name: "token read error: no push", tokenErr: errors.New("db read error")},
		{name: "token decrypt error: no push", ciphertext: storedCiphertext, decryptErr: errors.New("decrypt error")},
		{name: "push error: notification still completes", ciphertext: storedCiphertext, pushErr: errors.New("push error"), wantPush: true},
	}
	for _, event := range []string{"created", "cancelled"} {
		for _, tt := range tests {
			t.Run(event+"/"+tt.name, func(t *testing.T) {
				pusher := &fakeLinePusher{err: tt.pushErr}
				var gotClinicID uint64
				var gotCiphertext, gotToken string
				var mailTo []string
				repo := &mockLineSettingRepo{
					findTokenFn: func(_ context.Context, clinicID uint64) (string, error) {
						gotClinicID = clinicID
						return tt.ciphertext, tt.tokenErr
					},
				}
				decrypt := func(value string) (string, error) {
					gotCiphertext = value
					if tt.decryptErr != nil {
						return "", tt.decryptErr
					}
					return plainToken, nil
				}
				newMessenger := func(token string) LinePusher {
					gotToken = token
					return pusher
				}
				sendMail := func(_ context.Context, _ SMTPConfig, _, to string, _ []byte) error {
					mailTo = append(mailTo, to)
					return nil
				}
				cfg := &ReservationNotificationConfig{SMTPHost: "smtp.example.com", SMTPPort: "25", SMTPFrom: "from@example.com"}
				svc := NewReservationNotificationService(cfg, repo, decrypt, newMessenger, sendMail)
				appt := &model.Reservation{ID: 7, ClinicID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}
				customer := &model.LineCustomer{LineUserID: customerLineID, DisplayName: "山田太郎"}

				if event == "created" {
					svc.NotifyCreated(ctx, appt, customer)
				} else {
					svc.NotifyCancelled(ctx, appt, customer)
				}
				svc.Wait()

				assert.Equal(t, []string{"notify@example.com"}, mailTo, "clinic notification email must always be sent")
				assert.Equal(t, uint64(1), gotClinicID, "token must be read for the reservation's clinic")
				if !tt.wantPush {
					assert.Empty(t, pusher.pushes, "no LINE push expected")
					return
				}
				assert.Equal(t, storedCiphertext, gotCiphertext)
				assert.Equal(t, plainToken, gotToken)
				if assert.Len(t, pusher.pushes, 1) {
					assert.Equal(t, customerLineID, pusher.pushes[0].to)
					assert.Contains(t, pusher.pushes[0].text, "R-000007")
				}
			})
		}
	}
}

// 宛先は予約本人の LINE ユーザー ID だけである。LINE ID を持たない予約（顧客なし・LINE ID 空）では
// トークンの読取り・復号・Push のいずれも行わず、クリニック宛メールだけが残る。
func TestReservationNotificationService_LinePushNeedsReservingCustomerLineID(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		customer *model.LineCustomer
	}{
		{name: "nil customer"},
		{name: "empty LINE user id", customer: &model.LineCustomer{DisplayName: "山田太郎"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRead, messengerBuilt := false, false
			var mailTo []string
			repo := &mockLineSettingRepo{
				findTokenFn: func(_ context.Context, _ uint64) (string, error) {
					tokenRead = true
					return "ciphertext-aaaa", nil
				},
			}
			svc := NewReservationNotificationService(
				&ReservationNotificationConfig{SMTPHost: "smtp.example.com", SMTPPort: "25", SMTPFrom: "from@example.com"},
				repo,
				func(value string) (string, error) { return value, nil },
				func(string) LinePusher {
					messengerBuilt = true
					return &fakeLinePusher{}
				},
				func(_ context.Context, _ SMTPConfig, _, to string, _ []byte) error {
					mailTo = append(mailTo, to)
					return nil
				},
			)
			appt := &model.Reservation{ID: 8, ClinicID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}

			svc.NotifyCreated(ctx, appt, tt.customer)
			svc.Wait()

			assert.False(t, tokenRead, "token must not be read without a recipient")
			assert.False(t, messengerBuilt, "no LINE push without a recipient")
			assert.Equal(t, []string{"notify@example.com"}, mailTo)
		})
	}
}

// TestReservationNotificationService_SendEmail は SMTP 未設定時のスキップと、
// 接続失敗時のラップされたエラー返却を検証する（sendMail closure 注入後の R④ 形）。
// 本番 sendSMTPMail の auth 分岐（SMTPUser 設定時の smtp.PlainAuth 構築）は
// service/password_reset_service_test.go が実 sendSMTPMail で担保し、
// closure の SMTPConfig→smtpConfig フィールド対応は service/smtp_sender_adapter_test.go が検証する。
func TestReservationNotificationService_SendEmail(t *testing.T) {
	t.Run("skips send when SMTPHost is empty", func(t *testing.T) {
		svc := &reservationNotificationService{cfg: ReservationNotificationConfig{SMTPHost: ""}, sendMail: testDialSendMail}
		err := svc.sendEmail(context.Background(), "to@example.com", "subject", "body")
		assert.NoError(t, err)
	})

	t.Run("wraps error when SMTP connection fails without auth", func(t *testing.T) {
		svc := &reservationNotificationService{cfg: ReservationNotificationConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: "65533",
			SMTPFrom: "from@example.com",
		}, sendMail: testDialSendMail}
		err := svc.sendEmail(context.Background(), "to@example.com", "subject", "body")
		assert.Error(t, err)
	})

	t.Run("wraps error when SMTP connection fails with auth configured", func(t *testing.T) {
		svc := &reservationNotificationService{cfg: ReservationNotificationConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: "65533",
			SMTPUser: "user",
			SMTPPass: "pass",
			SMTPFrom: "from@example.com",
		}, sendMail: testDialSendMail}
		err := svc.sendEmail(context.Background(), "to@example.com", "subject", "body")
		assert.Error(t, err)
	})
}

// TestExtractPetNamesFromCustomerFields は appt.Pet 優先・customer_fields フォールバック・
// 不正JSON・空配列・type有無混在の各分岐を検証する。
func TestExtractPetNamesFromCustomerFields(t *testing.T) {
	tests := []struct {
		name string
		appt *model.Reservation
		want string
	}{
		{
			name: "appt.Pet takes priority over customer_fields",
			appt: &model.Reservation{
				Pet:            &model.Pet{Name: "ポチ"},
				CustomerFields: []byte(`{"pets": [{"name": "タマ", "type": "猫"}]}`),
			},
			want: "ポチ",
		},
		{
			name: "empty customer_fields returns empty string",
			appt: &model.Reservation{CustomerFields: nil},
			want: "",
		},
		{
			name: "invalid JSON returns empty string",
			appt: &model.Reservation{CustomerFields: []byte(`not-json`)},
			want: "",
		},
		{
			name: "empty pets array returns empty string",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": []}`)},
			want: "",
		},
		{
			name: "single pet with type",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": [{"name": "ポチ", "type": "犬"}]}`)},
			want: "ポチ(犬)",
		},
		{
			name: "single pet without type",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": [{"name": "ポチ"}]}`)},
			want: "ポチ",
		},
		{
			name: "multiple pets joined with japanese comma",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": [{"name": "ポチ", "type": "犬"}, {"name": "タマ"}]}`)},
			want: "ポチ(犬)、タマ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPetNamesFromCustomerFields(tt.appt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationHelpers(t *testing.T) {
	t.Run("reservationTypeDisplayName", func(t *testing.T) {
		assert.Equal(t, "", reservationTypeDisplayName(nil))
		assert.Equal(t, "disp", reservationTypeDisplayName(&model.ReservationType{ReservationDisplayName: "disp"}))
		assert.Equal(t, "short", reservationTypeDisplayName(&model.ReservationType{ShowShortName: true, ShortName: "short"}))
		assert.Equal(t, "name", reservationTypeDisplayName(&model.ReservationType{Name: "name"}))
	})

	t.Run("staffDisplayName", func(t *testing.T) {
		assert.Equal(t, "", staffDisplayName(nil))
		assert.Equal(t, "disp", staffDisplayName(&model.Staff{ReservationDisplayName: "disp"}))
		assert.Equal(t, "name", staffDisplayName(&model.Staff{Name: "name"}))
	})

	t.Run("customerDisplayName", func(t *testing.T) {
		assert.Equal(t, "不明", customerDisplayName(nil))
		assert.Equal(t, "real", customerDisplayName(&model.LineCustomer{RealName: "real", DisplayName: "disp"}))
		assert.Equal(t, "disp", customerDisplayName(&model.LineCustomer{DisplayName: "disp"}))
	})
}

// testDialSendMail は接続失敗パス検証用の実ダイヤル closure（本番は service 集約が
// service/smtp_sender.go の sendSMTPMail を注入する——本テストは到達不能ポートへの
// 実接続失敗のみを検証するため net.Dial 相当の最小実装で等価）。
func testDialSendMail(ctx context.Context, cfg SMTPConfig, _, _ string, _ []byte) error {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(cfg.Host, cfg.Port))
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
