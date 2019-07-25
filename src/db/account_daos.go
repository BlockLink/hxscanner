package db

import "time"

func SaveAccount(account *AccountEntity) error {
	now := time.Now()
	stmt, err := dbConn.Prepare("INSERT INTO public.account (owner_addr, account_name," +
		" created_at, updated_at)" +
		" VALUES (($1),($2),($3),($4))")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(account.OwnerAddr, account.AccountName, now.Unix(), now.Unix())
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func FindAccountByOwnerAddr(ownerAddr string) (result *AccountEntity, err error) {
	rows, err := dbConn.Query("SELECT id, owner_addr, account_name, created_at, updated_at" +
		" FROM public.account where owner_addr=$1", ownerAddr)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(AccountEntity)
		var createdAtUnix, updatedAtUnix int64
		err = rows.Scan(&result.Id, &result.OwnerAddr, &result.AccountName, &createdAtUnix, &updatedAtUnix)
		if err != nil {
			return
		}
		result.CreatedAt = time.Unix(createdAtUnix, 0)
		result.UpdatedAt = time.Unix(updatedAtUnix, 0)
		return
	} else {
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}
