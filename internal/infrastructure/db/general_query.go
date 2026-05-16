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
-- 最新の１件のみ取得
order_by effective_dt desc
limit 1
`
