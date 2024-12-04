package spray

var sprayedUsers = 0

func sprayCounter() {
	sprayedUsers += 1
}

var lockedAccounts = 0

func accountLocked() {
	lockedAccounts += 1
}

var checkedUsers = 0

func userChecked() {
	checkedUsers += 1
}