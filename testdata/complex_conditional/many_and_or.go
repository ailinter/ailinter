package main

func IsEligible(user *User) bool {
	if user.Age > 18 && user.HasVerifiedEmail && !user.IsBanned && (user.Subscription == "premium" || user.PurchaseTotal > 1000 || user.Referrals > 5) {
		return true
	}
	return false
}
