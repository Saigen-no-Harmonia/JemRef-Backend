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
-- 最新の1件のみを取得
order by effective_dt desc
limit 1
`
