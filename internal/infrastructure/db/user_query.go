package db

const createUserQuery = `
insert into m_users (
	PUBLIC_ID,
	FIREBASE_ID,
	EMAIL,
	PRIVACY_POLICY_AGREED_DT,
	PRIVACY_POLICY_VERSION,
	TERMS_AGREED_DT,
	TERMS_VERSION,
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
