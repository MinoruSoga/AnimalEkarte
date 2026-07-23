package handler

import (
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
)

type clinicSettingsResponse = clinicdomain.ClinicSettingsResponse
type closingSpecialPeriodResponse = clinicdomain.ClosingSpecialPeriodResponse
type closingSettingsFullResponse = clinicdomain.ClosingSettingsFullResponse

func toClinicSettingsResponse(settings *model.ClinicSettings) clinicSettingsResponse {
	return clinicdomain.ToClinicSettingsResponse(settings)
}

func toClosingSettingsFullResponse(
	settings *model.ClinicSettings,
	periods []model.ClosingSpecialPeriod,
) closingSettingsFullResponse {
	return clinicdomain.ToClosingSettingsFullResponse(settings, periods)
}

func toClosingSpecialPeriodResponse(period *model.ClosingSpecialPeriod) closingSpecialPeriodResponse {
	return clinicdomain.ToClosingSpecialPeriodResponse(period)
}
