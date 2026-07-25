package db

const selectPolicyByIdQuery = `
select 
	ID,
	VERSION,
	NAME,
	CONTENT,
	EFFECTIVE_DT
from m_policies
where 
	DEL_FLG=0
and
	ID = ?
-- 発効済みかつ最新の1件を取得
and
	EFFECTIVE_DT <= CURDATE()
order by 
	EFFECTIVE_DT desc,
	VERSION desc
limit 1
`
const selectPolicyByPrimaryKeyQuery = `
select 
	ID,
	VERSION,
	NAME,
	CONTENT,
	EFFECTIVE_DT
from m_policies
where 
	DEL_FLG = 0
and
	ID = ?
and
	VERSION = ?
`
