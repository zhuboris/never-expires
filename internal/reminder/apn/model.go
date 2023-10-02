package apn

type notificationData struct {
	DeviceToken             string
	ClosestExpiringItemName string
	ExpiringSoonItemsCount  int
}
