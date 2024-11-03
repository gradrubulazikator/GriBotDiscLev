package main

import (
	"fmt"
	"log"
	"strings"
	"time"
        "strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const telegramBotToken = "7392721358:AAG2pKYglfGuNcTm7xAqicQFCoRsajSZUGs"

var tasks = make(map[int64][]string) // хранение задач по chat ID

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Запуск ежедневных напоминаний
	go dailyReminder(bot)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в GriBotDiscLev! Используйте /addtask для добавления задачи, /tasks для просмотра задач или /done для отметки задачи как выполненной.")
			bot.Send(msg)

		case "/tasks":
			listTasks(update.Message.Chat.ID, bot)

		case "/help":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команды:\n/addtask [описание] - добавить задачу\n/tasks - список задач\n/done [номер] - отметить задачу как выполненную\n/help - показать команды")
			bot.Send(msg)

		default:
			if strings.HasPrefix(update.Message.Text, "/addtask") {
				addTask(update.Message.Chat.ID, update.Message.Text, bot)
			} else if strings.HasPrefix(update.Message.Text, "/done") {
				markTaskAsDone(update.Message.Chat.ID, update.Message.Text, bot)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для просмотра списка команд.")
				bot.Send(msg)
			}
		}
	}
}

// addTask добавляет новую задачу
func addTask(chatID int64, text string, bot *tgbotapi.BotAPI) {
	task := strings.TrimPrefix(text, "/addtask ")
	if task == "" {
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, добавьте описание задачи после команды /addtask.")
		bot.Send(msg)
		return
	}
	tasks[chatID] = append(tasks[chatID], task)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Задача '%s' добавлена!", task))
	bot.Send(msg)
}

// listTasks отправляет список задач пользователю
func listTasks(chatID int64, bot *tgbotapi.BotAPI) {
	if len(tasks[chatID]) == 0 {
		msg := tgbotapi.NewMessage(chatID, "У вас нет активных задач.")
		bot.Send(msg)
		return
	}

	var taskList string
	for i, task := range tasks[chatID] {
		taskList += fmt.Sprintf("%d. %s\n", i+1, task)
	}
	msg := tgbotapi.NewMessage(chatID, "Ваши задачи:\n"+taskList)
	bot.Send(msg)
}

// markTaskAsDone отмечает задачу как выполненную
func markTaskAsDone(chatID int64, text string, bot *tgbotapi.BotAPI) {
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, укажите номер задачи после команды /done.")
		bot.Send(msg)
		return
	}

	taskNumber := parts[1]
	taskIndex, err := strconv.Atoi(taskNumber)
	if err != nil || taskIndex < 1 || taskIndex > len(tasks[chatID]) {
		msg := tgbotapi.NewMessage(chatID, "Некорректный номер задачи.")
		bot.Send(msg)
		return
	}

	// Удаление выполненной задачи
	task := tasks[chatID][taskIndex-1]
	tasks[chatID] = append(tasks[chatID][:taskIndex-1], tasks[chatID][taskIndex:]...)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Задача '%s' отмечена как выполненная!", task))
	bot.Send(msg)
}

// dailyReminder отправляет напоминания о невыполненных задачах ежедневно
func dailyReminder(bot *tgbotapi.BotAPI) {
	for {
		now := time.Now()
		nextReminder := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())
		time.Sleep(nextReminder.Sub(now))

		for chatID, userTasks := range tasks {
			if len(userTasks) > 0 {
				var taskList string
				for i, task := range userTasks {
					taskList += fmt.Sprintf("%d. %s\n", i+1, task)
				}
				msg := tgbotapi.NewMessage(chatID, "Напоминание о ваших задачах:\n"+taskList)
				bot.Send(msg)
			}
		}
	}
}

