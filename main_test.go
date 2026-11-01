package main

import "testing"

func checkCommands(t *testing.T, input, want string) {
	t.Helper()
	if got := runCommands(input); got != want {
		t.Fatalf("runCommands(%q):\nполучено: %q\nожидалось: %q", input, got, want)
	}
}

// Обязательные команды базового сценария: catalog, stock, cart, add, remove,
// promo SALE10, promo off и checkout. Проверяется их работа в одной сессии.
func TestBasicScenario(t *testing.T) {
	input := `catalog
stock coffee
cart
add coffee 2
add mug 1
cart
remove mug 1
cart
add mug 1
promo SALE10
cart
promo off
cart
promo SALE10
checkout
stock coffee
stock mug
cart
checkout
`
	want := `Каталог:
tea - 200 ₽ (10 шт.)
coffee - 500 ₽ (5 шт.)
mug - 700 ₽ (3 шт.)
На складе: coffee - 5 шт.
Корзина пуста
Добавлено: coffee x2
Добавлено: mug x1
coffee x2 - 1000 ₽
mug x1 - 700 ₽
Товары: 1700 ₽
Скидка: 0 ₽
Доставка: 0 ₽
Итого: 1700 ₽
Удалено: mug x1
coffee x2 - 1000 ₽
Товары: 1000 ₽
Скидка: 0 ₽
Доставка: 200 ₽
Итого: 1200 ₽
Добавлено: mug x1
Промокод SALE10 применён
coffee x2 - 1000 ₽
mug x1 - 700 ₽
Товары: 1700 ₽
Скидка: 170 ₽
Доставка: 0 ₽
Итого: 1530 ₽
Промокод отключён
coffee x2 - 1000 ₽
mug x1 - 700 ₽
Товары: 1700 ₽
Скидка: 0 ₽
Доставка: 0 ₽
Итого: 1700 ₽
Промокод SALE10 применён
Заказ №1 оформлен. К оплате: 1530 ₽
На складе: coffee - 3 шт.
На складе: mug - 2 шт.
Корзина пуста
Ошибка: корзина пуста
`
	checkCommands(t, input, want)
}

// Для каждой известной команды число аргументов обязательно проверяется.
func TestArgumentCount(t *testing.T) {
	cases := []string{
		"catalog extra", "stock", "stock tea extra",
		"add", "add tea", "add tea 2 extra",
		"remove", "remove tea", "remove tea 1 extra",
		"cart extra", "promo", "promo SALE10 extra", "checkout extra",
	}
	for _, command := range cases {
		t.Run(command, func(t *testing.T) {
			checkCommands(t, command+"\n", "Ошибка: неверная команда\n")
		})
	}
}

func TestErrorsPreserveState(t *testing.T) {
	input := "add tea 1\nadd tea 0\nadd banana 1\nadd tea 2 extra\nremove tea 2\ncart\n"
	want := "Добавлено: tea x1\n" +
		"Ошибка: количество должно быть положительным\n" +
		"Ошибка: неизвестный товар\n" +
		"Ошибка: неверная команда\n" +
		"Ошибка: недостаточно товара в корзине\n" +
		"tea x1 - 200 ₽\nТовары: 200 ₽\nСкидка: 0 ₽\nДоставка: 200 ₽\nИтого: 400 ₽\n"
	checkCommands(t, input, want)
}

func TestNewCallStartsNewSession(t *testing.T) {
	checkCommands(t, "add tea 1\n", "Добавлено: tea x1\n")
	checkCommands(t, "cart\n", "Корзина пуста\n")
	checkCommands(t, "cart", "Корзина пуста\n")
}

func TestBlankLinesCommentsAndSpaces(t *testing.T) {
	checkCommands(t, "", "")
	checkCommands(t, "\n  # комментарий\n  stock  \t tea  \n\n", "На складе: tea - 10 шт.\n")
	checkCommands(t, "unknown\n", "Ошибка: неизвестная команда\n")
}

func TestPromoThresholdAndDelivery(t *testing.T) {
	checkCommands(t, "promo SALE10\nadd tea 4\ncart\nadd tea 1\ncart\n",
		"Промокод SALE10 применён\nДобавлено: tea x4\n"+
			"tea x4 - 800 ₽\nТовары: 800 ₽\nСкидка: 0 ₽\nДоставка: 200 ₽\nИтого: 1000 ₽\n"+
			"Добавлено: tea x1\n"+
			"tea x5 - 1000 ₽\nТовары: 1000 ₽\nСкидка: 100 ₽\nДоставка: 200 ₽\nИтого: 1100 ₽\n")

	checkCommands(t, "add tea 1\nadd mug 2\ncart\npromo SALE10\ncart\n",
		"Добавлено: tea x1\nДобавлено: mug x2\n"+
			"tea x1 - 200 ₽\nmug x2 - 1400 ₽\nТовары: 1600 ₽\nСкидка: 0 ₽\nДоставка: 0 ₽\nИтого: 1600 ₽\n"+
			"Промокод SALE10 применён\n"+
			"tea x1 - 200 ₽\nmug x2 - 1400 ₽\nТовары: 1600 ₽\nСкидка: 160 ₽\nДоставка: 200 ₽\nИтого: 1640 ₽\n")
}

func TestStockLimitAndSecondOrder(t *testing.T) {
	checkCommands(t, "add coffee 4\nadd coffee 2\npromo SALE10\ncheckout\nstock coffee\nadd coffee 2\nadd coffee 1\ncheckout\nstock coffee\n",
		"Добавлено: coffee x4\n"+
			"Ошибка: недостаточно товара на складе\n"+
			"Промокод SALE10 применён\n"+
			"Заказ №1 оформлен. К оплате: 1800 ₽\n"+
			"На складе: coffee - 1 шт.\n"+
			"Ошибка: недостаточно товара на складе\n"+
			"Добавлено: coffee x1\n"+
			"Заказ №2 оформлен. К оплате: 700 ₽\n"+
			"На складе: coffee - 0 шт.\n")
}

func TestInvalidValues(t *testing.T) {
	checkCommands(t, "add tea abc\nadd tea -2\nremove tea abc\nstock banana\npromo NOPE\ncart\n",
		"Ошибка: некорректное количество\n"+
			"Ошибка: количество должно быть положительным\n"+
			"Ошибка: некорректное количество\n"+
			"Ошибка: неизвестный товар\n"+
			"Ошибка: неизвестный промокод\n"+
			"Корзина пуста\n")
}
