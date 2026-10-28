package practice

import (
	"errors"
	"reflect"
	"testing"
)

func TestSMSValidationPreservesHistory(t *testing.T) {
	sms := &SMSNotifier{
		BaseNotifier: BaseNotifier{Channel: "sms"},
		Messages:     []string{"[sms] Старое сообщение"},
	}
	if err := sms.Send(""); err != ErrEmptyMessage {
		t.Fatalf("пустой текст должен проверяться раньше номера: %v", err)
	}
	var target *SendError
	if err := sms.Send("Заказ готов"); !errors.As(err, &target) {
		t.Fatalf("без номера нужен *SendError: %v", err)
	}
	if target.Channel != "sms" || target.Reason != "не задан номер телефона" {
		t.Fatalf("неверная ошибка настройки: %+v", target)
	}
	if !reflect.DeepEqual(sms.Messages, []string{"[sms] Старое сообщение"}) {
		t.Fatalf("ошибки изменили историю: %v", sms.Messages)
	}
}

func TestSMSDispatchers(t *testing.T) {
	cases := []struct {
		name string
		send func([]Notifier, string) []error
	}{
		{"Dispatch", func(all []Notifier, message string) []error {
			if err := Dispatch(all, message); err != nil {
				return []error{err}
			}
			return nil
		}},
		{"SafeDispatch", SafeDispatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			email, push := newEmail(), newPush()
			sms := &SMSNotifier{
				BaseNotifier: BaseNotifier{Channel: "sms"},
				PhoneNumber:  "demo-number",
			}
			all := []Notifier{email, push, sms}
			for _, message := range []string{"Заказ готов", " "} {
				if failures := tc.send(all, message); len(failures) != 0 {
					t.Fatalf("рассылка через три канала: %v", failures)
				}
			}
			for channel, messages := range map[string][]string{
				"email": email.Messages,
				"push":  push.Messages,
				"sms":   sms.Messages,
			} {
				want := []string{"[" + channel + "] Заказ готов", "[" + channel + "]  "}
				if !reflect.DeepEqual(messages, want) {
					t.Fatalf("история %s: %v, ожидается %v", channel, messages, want)
				}
			}
		})
	}
}

func TestSMSAfterPanic(t *testing.T) {
	sms := &SMSNotifier{
		BaseNotifier: BaseNotifier{Channel: "sms"},
		PhoneNumber:  "demo-number",
	}
	failures := SafeDispatch([]Notifier{&BrokenNotifier{}, sms}, "Заказ готов")
	if len(failures) != 1 || !errors.Is(failures[0], ErrNotifierPanic) {
		t.Fatalf("ожидается одна ошибка паники: %v", failures)
	}
	if !reflect.DeepEqual(sms.Messages, []string{"[sms] Заказ готов"}) {
		t.Fatalf("SMS после паники не получил сообщение: %v", sms.Messages)
	}
}
