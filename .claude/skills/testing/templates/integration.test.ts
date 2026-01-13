/**
 * Integration Test Template
 *
 * 複数コンポーネント間の連携テスト
 * データベース/API を含むテスト
 */

import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { PrismaClient } from '@prisma/client';
// import { app } from '../src/app';
// import request from 'supertest';

const prisma = new PrismaClient();

describe('Integration: UserService', () => {
  // テストスイート全体の前処理
  beforeAll(async () => {
    // データベース接続
    await prisma.$connect();
  });

  // テストスイート全体の後処理
  afterAll(async () => {
    // クリーンアップとデータベース切断
    await prisma.user.deleteMany({});
    await prisma.$disconnect();
  });

  // 各テスト前のセットアップ
  beforeEach(async () => {
    // テストデータのリセット
    await prisma.user.deleteMany({});
  });

  describe('createUser', () => {
    it('should create user and persist to database', async () => {
      // Arrange
      const userData = {
        email: 'test@example.com',
        name: 'Test User',
      };

      // Act
      const user = await prisma.user.create({ data: userData });

      // Assert
      expect(user.id).toBeDefined();
      expect(user.email).toBe(userData.email);

      // Verify persistence
      const savedUser = await prisma.user.findUnique({
        where: { id: user.id },
      });
      expect(savedUser).not.toBeNull();
    });

    it('should reject duplicate email', async () => {
      // Arrange
      const userData = { email: 'duplicate@example.com', name: 'User 1' };
      await prisma.user.create({ data: userData });

      // Act & Assert
      await expect(
        prisma.user.create({ data: { ...userData, name: 'User 2' } })
      ).rejects.toThrow();
    });
  });

  describe('getUserWithPosts', () => {
    it('should return user with related posts', async () => {
      // Arrange
      const user = await prisma.user.create({
        data: {
          email: 'author@example.com',
          name: 'Author',
          posts: {
            create: [
              { title: 'Post 1', content: 'Content 1' },
              { title: 'Post 2', content: 'Content 2' },
            ],
          },
        },
      });

      // Act
      const result = await prisma.user.findUnique({
        where: { id: user.id },
        include: { posts: true },
      });

      // Assert
      expect(result?.posts).toHaveLength(2);
      expect(result?.posts[0].title).toBe('Post 1');
    });
  });
});

describe('Integration: API Endpoints', () => {
  describe('GET /api/users', () => {
    it('should return list of users', async () => {
      // Arrange
      await prisma.user.createMany({
        data: [
          { email: 'user1@example.com', name: 'User 1' },
          { email: 'user2@example.com', name: 'User 2' },
        ],
      });

      // Act
      // const response = await request(app).get('/api/users');

      // Assert
      // expect(response.status).toBe(200);
      // expect(response.body).toHaveLength(2);
    });

    it('should return 401 without authentication', async () => {
      // Act
      // const response = await request(app).get('/api/users');

      // Assert
      // expect(response.status).toBe(401);
    });
  });

  describe('POST /api/users', () => {
    it('should create user and return 201', async () => {
      // Arrange
      const newUser = { email: 'new@example.com', name: 'New User' };

      // Act
      // const response = await request(app)
      //   .post('/api/users')
      //   .send(newUser)
      //   .set('Authorization', 'Bearer valid-token');

      // Assert
      // expect(response.status).toBe(201);
      // expect(response.body.email).toBe(newUser.email);
    });

    it('should return 400 for invalid data', async () => {
      // Arrange
      const invalidUser = { email: 'invalid-email' };

      // Act
      // const response = await request(app)
      //   .post('/api/users')
      //   .send(invalidUser);

      // Assert
      // expect(response.status).toBe(400);
    });
  });
});

/**
 * Integration Test チェックリスト:
 * - [ ] データベース操作の永続化確認
 * - [ ] トランザクションの整合性
 * - [ ] 関連データの取得（JOIN相当）
 * - [ ] API レスポンスステータス
 * - [ ] 認証/認可フロー
 * - [ ] エラーレスポンス形式
 */
