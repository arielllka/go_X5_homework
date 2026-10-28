package main

import (
	"errors"
	"fmt"

	"lesson5/student/practice"
)

func main() {
	const message = "Заказ готов"
	email := &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "student@example.com",
	}
	push := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "phone-1",
	}
	sms := &practice.SMSNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "sms"},
		PhoneNumber:  "demo-number-1",
	}
	all := []practice.Notifier{email, push, sms}

	fmt.Println("1. Успешная рассылка")
	if err := practice.Dispatch(all, message); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("Все три канала получили сообщение, ошибок нет.")
	}
	fmt.Printf("История email: %q\nИстория push: %q\n", email.Messages, push.Messages)
	fmt.Printf("История SMS: %q\n", sms.Messages)

	fmt.Println("\n2. Пустой текст")
	if err := practice.Dispatch(all, ""); errors.Is(err, practice.ErrEmptyMessage) {
		fmt.Printf("errors.Is распознал ErrEmptyMessage: %v\n", err)
	} else {
		fmt.Printf("Неожиданный результат: %v\n", err)
	}
	fmt.Printf("История email: %q\nИстория push: %q\n", email.Messages, push.Messages)
	fmt.Printf("История SMS: %q\n", sms.Messages)

	fmt.Println("\n3. Нет настройки push")
	unconfiguredPush := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
	}
	err := practice.Dispatch([]practice.Notifier{unconfiguredPush}, message)
	var sendErr *practice.SendError
	if errors.As(err, &sendErr) {
		fmt.Printf("errors.As извлёк *SendError: Channel=%q, Reason=%q\n", sendErr.Channel, sendErr.Reason)
	} else {
		fmt.Printf("Неожиданный результат: %v\n", err)
	}
	fmt.Printf("История push без настройки: %q\n", unconfiguredPush.Messages)

	fmt.Println("\n4. Продолжение после паники")
	broken := &practice.BrokenNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "broken"},
	}
	nextPush := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "phone-2",
	}
	nextSMS := &practice.SMSNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "sms"},
		PhoneNumber:  "demo-number-2",
	}
	failures := practice.SafeDispatch([]practice.Notifier{broken, nextPush, nextSMS}, message)
	fmt.Printf("Ошибок: %d\n", len(failures))
	for _, err := range failures {
		if errors.Is(err, practice.ErrNotifierPanic) {
			fmt.Printf("errors.Is распознал ErrNotifierPanic: %v\n", err)
		} else {
			fmt.Printf("Ошибка: %v\n", err)
		}
	}
	fmt.Printf("История следующего push: %q\n", nextPush.Messages)
	fmt.Printf("История следующего SMS: %q\n", nextSMS.Messages)
}
