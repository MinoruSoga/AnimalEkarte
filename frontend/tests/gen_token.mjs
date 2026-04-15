import jwt from 'jsonwebtoken';
const secret = 'dev-secret-change-me';
const claims = {
  user_id: '1',
  clinic_id: '1',
  is_system_admin: true,
  clinic_ids: [1],
  iat: Math.floor(Date.now() / 1000),
  exp: Math.floor(Date.now() / 1000) + (60 * 60)
};
const token = jwt.sign(claims, secret);
console.log(token);
