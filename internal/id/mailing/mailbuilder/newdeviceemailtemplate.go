package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type newDeviceTemplateInput struct {
	Subject      string
	Header       string
	Body         string
	TimeRow      string
	Time         string
	DeviceRow    string
	Device       string
	IPAddressRow string
	IP           string
	Annotation   string
}

func (b Builder) newNewDeviceTemplateInput(notificationData NotificationData, language lang.Language) (newDeviceTemplateInput, error) {
	content := b.localesDict.NewDevice
	input, err := b.newMessageEmailTemplateInput(content.messageEmailContent, language)
	if err != nil {
		return newDeviceTemplateInput{}, err
	}

	timeRow, err := content.TimeRow.requestedOrDefaultValue(language)
	if err != nil {
		return newDeviceTemplateInput{}, err
	}

	deviceRow, err := content.DeviceRow.requestedOrDefaultValue(language)
	if err != nil {
		return newDeviceTemplateInput{}, err
	}

	ipRow, err := content.IPAddressRow.requestedOrDefaultValue(language)
	if err != nil {
		return newDeviceTemplateInput{}, err
	}

	return newDeviceTemplateInput{
		Subject:      input.subject,
		Header:       input.header,
		Body:         input.body,
		TimeRow:      timeRow,
		Time:         notificationData.formattedTime(),
		DeviceRow:    deviceRow,
		Device:       notificationData.device(),
		IPAddressRow: ipRow,
		IP:           notificationData.ipWithLocation(),
		Annotation:   input.annotation,
	}, nil
}
