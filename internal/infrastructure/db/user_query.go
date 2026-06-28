package db

const createUserQuery = `
insert into m_users (
	PUBLIC_ID,
	FIREBASE_ID,
	EMAIL,
	TERMS_AGREED_DT,
	TERMS_VERSION,
	PRIVACY_POLICY_AGREED_DT,
	PRIVACY_POLICY_VERSION,
	INS_PG,
	INS_ID,
	UPD_PG,
	UPD_ID
)
values (
	?,
	?,
	?,
	?,
	?,
	?,
	?,
	'system',
	'MEM-API-002',
	'system',
	'MEM-API-002'
)
`

const deleteUserQuery = `
UPDATE m_users 
SET
  EMAIL = CONCAT('__deleted__', EMAIL, '__', NOW()),
  DEL_FLG = 1,
  DELETED_AT = NOW(),
  UPD_PG = 'MEM-API-004',
  UPD_ID = 'system',
  UPD_DT = NOW()
WHERE
  ID = ?
`

const selectUserByUidQuery = `
	select 
		ID,
		PUBLIC_ID,
		FIREBASE_ID,
		EMAIL,
		TERMS_AGREED_DT,
		TERMS_VERSION,
		PRIVACY_POLICY_AGREED_DT,
		PRIVACY_POLICY_VERSION,
		DELETED_AT
	from
		m_users
	where
		ID = ?
	and
		DEL_FLG = '0'
`

const selectUserByUFirebaseUidQuery = `
	select 
		ID,
		PUBLIC_ID,
		FIREBASE_ID,
		EMAIL,
		TERMS_AGREED_DT,
		TERMS_VERSION,
		PRIVACY_POLICY_AGREED_DT,
		PRIVACY_POLICY_VERSION,
		DELETED_AT
	from
		m_users
	where
		FIREBASE_ID = ?
	and
		DEL_FLG = '0'
`
