package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	input := strings.Join(os.Args[1:], "\n")
	if len(os.Args) > 1 && os.Args[1] == "-file" {
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Использование: go run . -file путь")
			os.Exit(1)
		}
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		input = string(data)
	}
	fmt.Print(runCommands(input))
}

type product struct {
	code              string
	price, stock, qty int
}

func runCommands(input string) string {
	ps := []product{{"tea", 200, 10, 0}, {"coffee", 500, 5, 0}, {"mug", 700, 3, 0}}
	promo := false
	order := 1
	var out strings.Builder
	say := func(s string) { out.WriteString(s + "\n") }
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		a := strings.Fields(line)
		count := 0
		switch a[0] {
		case "catalog", "cart", "checkout":
			count = 1
		case "stock", "promo":
			count = 2
		case "add", "remove":
			count = 3
		default:
			say("Ошибка: неизвестная команда")
			continue
		}
		if len(a) != count {
			say("Ошибка: неверная команда")
			continue
		}
		switch a[0] {
		case "catalog":
			say("Каталог:")
			for _, p := range ps {
				say(fmt.Sprintf("%s - %d ₽ (%d шт.)", p.code, p.price, p.stock))
			}
		case "stock", "add", "remove":
			idx := -1
			for i, p := range ps {
				if p.code == a[1] {
					idx = i
					break
				}
			}
			if idx < 0 {
				say("Ошибка: неизвестный товар")
				continue
			}
			p := &ps[idx]
			if a[0] == "stock" {
				say(fmt.Sprintf("На складе: %s - %d шт.", p.code, p.stock))
				continue
			}
			n, err := strconv.Atoi(a[2])
			if err != nil {
				say("Ошибка: некорректное количество")
				continue
			}
			if n <= 0 {
				say("Ошибка: количество должно быть положительным")
				continue
			}
			if a[0] == "add" {
				if n > p.stock-p.qty {
					say("Ошибка: недостаточно товара на складе")
					continue
				}
				p.qty += n
				say(fmt.Sprintf("Добавлено: %s x%d", p.code, n))
			} else {
				if n > p.qty {
					say("Ошибка: недостаточно товара в корзине")
					continue
				}
				p.qty -= n
				say(fmt.Sprintf("Удалено: %s x%d", p.code, n))
			}
		case "promo":
			switch a[1] {
			case "SALE10":
				promo = true
				say("Промокод SALE10 применён")
			case "off":
				promo = false
				say("Промокод отключён")
			default:
				say("Ошибка: неизвестный промокод")
			}
		case "cart", "checkout":
			sum := 0
			for _, p := range ps {
				sum += p.price * p.qty
			}
			if sum == 0 {
				if a[0] == "cart" {
					say("Корзина пуста")
				} else {
					say("Ошибка: корзина пуста")
				}
				continue
			}
			discount := 0
			if promo && sum >= 1000 {
				discount = sum / 10
			}
			delivery := 0
			if sum-discount < 1500 {
				delivery = 200
			}
			total := sum - discount + delivery
			if a[0] == "cart" {
				for _, p := range ps {
					if p.qty > 0 {
						say(fmt.Sprintf("%s x%d - %d ₽", p.code, p.qty, p.price*p.qty))
					}
				}
				say(fmt.Sprintf("Товары: %d ₽\nСкидка: %d ₽\nДоставка: %d ₽\nИтого: %d ₽", sum, discount, delivery, total))
			} else {
				say(fmt.Sprintf("Заказ №%d оформлен. К оплате: %d ₽", order, total))
				order++
				for i := range ps {
					ps[i].stock -= ps[i].qty
					ps[i].qty = 0
				}
				promo = false
			}
		}
	}
	return out.String()
}
