package bot

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/node_tooling/Celo/cmd"
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/synced"),
		tgbotapi.NewKeyboardButton("/balance"),
		tgbotapi.NewKeyboardButton("/status"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/score"),
		tgbotapi.NewKeyboardButton("/signing"),
		tgbotapi.NewKeyboardButton("/exchange_rate"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/lockgold"),
		tgbotapi.NewKeyboardButton("/exchange"),
		tgbotapi.NewKeyboardButton("/vote"),
	),
	tgbotapi.NewKeyboardButtonRow(
		// tgbotapi.NewKeyboardButton("/exchange"),
		// tgbotapi.NewKeyboardButton("/exchange_rate"),
		tgbotapi.NewKeyboardButton("/close"),
	),
)

var lockGoldKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Validator Group All", "valGrLockGold"),
		tgbotapi.NewInlineKeyboardButtonData("Validator All", "valLockGold"),
	),
	// tgbotapi.NewInlineKeyboardRow(
	// 	tgbotapi.NewInlineKeyboardButtonData("Validator Amount", "valAmount"),
	// 	tgbotapi.NewInlineKeyboardButtonData("Validator Group Amount", "valGrAmount"),
	// ),
)

var exchangeUsdKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Validator Group 25%", "valGrOneForthUsd"),
		tgbotapi.NewInlineKeyboardButtonData("Validator Group 50%", "valGrHalfUsd"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Validator Group 75%", "valGrFourThirdsUsd"),
		tgbotapi.NewInlineKeyboardButtonData("Validator Group All", "valGrAllUsd"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Validator 25%", "valOneFourthUsd"),
		tgbotapi.NewInlineKeyboardButtonData("Validator 50%", "valHalfUsd"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Validator 75%", "valFourThirdsUsd"),
		tgbotapi.NewInlineKeyboardButtonData("Validator All", "valAllUsd"),
	),
)

var electionVoteKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Validator Group All", "valGrVote"),
		tgbotapi.NewInlineKeyboardButtonData("Validator All", "valVote"),
	),
)

type perform interface {
	getBalance(msg tgbotapi.MessageConfig)
	lockGold(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig)
	exchanegUSDToGold(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig, perct uint16)
	vote(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig)
}

type balance struct {
	gold       string
	usd        string
	lockedGold string
	nonVoting  string
	total      string
}

type validator struct {
	balance
}

type validatorGr struct {
	balance
}

// Run instantiates the bot
func Run() {
	botToken := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		// ignore any non-Message Updates
		// if update.Message == nil {
		// 	continue
		// }

		// ignore any non-command Messages
		// if !update.Message.IsCommand() {
		// 	continue
		// }

		if update.CallbackQuery != nil {
			// msg := update.CallbackQuery.Message
			data := update.CallbackQuery.Data
			chatID := update.CallbackQuery.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "")
			msg.ParseMode = "Markdown"
			switch data {
			case "valGrLockGold":
				var valGr validatorGr
				UpdateBalance(&valGr, msg) // update balance before locking
				lockGoldRun(&valGr, bot, msg)
				UpdateBalance(&valGr, msg) // update balance after locking
				msgPiece := `gold: ` + valGr.balance.gold + "\n" + `lockedGold: ` + valGr.balance.lockedGold
				msg.Text = boldText("Validator Group Balance After Locking") + "\n\n" + msgPiece
				break
			case "valLockGold":
				var val validator
				UpdateBalance(&val, msg) // update balance before locking
				lockGoldRun(&val, bot, msg)
				UpdateBalance(&val, msg) // update balance after locking
				msgPiece := `gold: ` + val.balance.gold + "\n" + `lockedGold: ` + val.balance.lockedGold
				msg.Text = boldText("Validator Balance After Locking") + "\n\n" + msgPiece
				break
			case "valAmount":
				msg.Text = "Locking a specific amount from validator was requested"
				break
			case "valGrAmount":
				msg.Text = "Locking a specific amount from validator group was requested"
				break
			case "valGrAllUsd":
				var valGr validatorGr
				UpdateBalance(&valGr, msg) // update balance before exchange
				exchangeUSDToGoldRun(&valGr, bot, msg, 100)
				UpdateBalance(&valGr, msg) // update balance after exchange
				msgPiece := `gold: ` + valGr.balance.gold + "\n" + `usd: ` + valGr.balance.usd
				msg.Text = boldText("Validator Group Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valGrHalfUsd":
				var valGr validatorGr
				UpdateBalance(&valGr, msg) // update balance before exchange
				exchangeUSDToGoldRun(&valGr, bot, msg, 50)
				UpdateBalance(&valGr, msg) // update balance after exchange
				msgPiece := `gold: ` + valGr.balance.gold + "\n" + `usd: ` + valGr.balance.usd
				msg.Text = boldText("Validator Group Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valGrOneForthUsd":
				var valGr validatorGr
				UpdateBalance(&valGr, msg) // update balance before exchange
				exchangeUSDToGoldRun(&valGr, bot, msg, 25)
				UpdateBalance(&valGr, msg) // update balance after exchange
				msgPiece := `gold: ` + valGr.balance.gold + "\n" + `usd: ` + valGr.balance.usd
				msg.Text = boldText("Validator Group Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valGrFourThirdsUsd":
				var valGr validatorGr
				UpdateBalance(&valGr, msg) // update balance before exchange
				exchangeUSDToGoldRun(&valGr, bot, msg, 75)
				UpdateBalance(&valGr, msg) // update balance after exchange
				msgPiece := `gold: ` + valGr.balance.gold + "\n" + `usd: ` + valGr.balance.usd
				msg.Text = boldText("Validator Group Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valAllUsd":
				var val validator
				UpdateBalance(&val, msg) // update balance before exchange
				exchangeUSDToGoldRun(&val, bot, msg, 100)
				UpdateBalance(&val, msg) // update balance after exchange
				msgPiece := `gold: ` + val.balance.gold + "\n" + `usd: ` + val.balance.usd
				msg.Text = boldText("Validator Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valHalfUsd":
				var val validator
				UpdateBalance(&val, msg) // update balance before exchange
				exchangeUSDToGoldRun(&val, bot, msg, 50)
				UpdateBalance(&val, msg) // update balance after exchange
				msgPiece := `gold: ` + val.balance.gold + "\n" + `usd: ` + val.balance.usd
				msg.Text = boldText("Validator Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valOneForthUsd":
				var val validator
				UpdateBalance(&val, msg) // update balance before exchange
				exchangeUSDToGoldRun(&val, bot, msg, 25)
				UpdateBalance(&val, msg) // update balance after exchange
				msgPiece := `gold: ` + val.balance.gold + "\n" + `usd: ` + val.balance.usd
				msg.Text = boldText("Validator Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valFourThirdsUsd":
				var val validator
				UpdateBalance(&val, msg) // update balance before exchange
				exchangeUSDToGoldRun(&val, bot, msg, 75)
				UpdateBalance(&val, msg) // update balance after exchange
				msgPiece := `gold: ` + val.balance.gold + "\n" + `usd: ` + val.balance.usd
				msg.Text = boldText("Validator Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valGrVote":
				var valGr validatorGr
				UpdateBalance(&valGr, msg) // update balance before voting
				voteRun(&valGr, bot, msg)
				UpdateBalance(&valGr, msg) // update balance after voting
				msgPiece := `Non-voting: ` + valGr.balance.nonVoting
				msg.Text = boldText("Validator Group Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "valVote":
				var val validator
				UpdateBalance(&val, msg) // update balance before voting
				voteRun(&val, bot, msg)
				UpdateBalance(&val, msg) // update balance after voting
				msgPiece := `Non-voting: ` + val.balance.nonVoting
				msg.Text = boldText("Validator Balance After Exhchange") + "\n\n" + msgPiece
				break
			case "cancel":
				msg.Text = "Back to back menu"
				break
			}

			// send final message out
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue
		}
		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		chatID := update.Message.Chat.ID
		msg := tgbotapi.NewMessage(chatID, "")
		msg.ParseMode = "Markdown"
		// msg.ReplyToMessageID = update.Message.MessageID
		// bot.Send(msg)
		switch update.Message.Command() {
		case "help":
			msg.Text = "type /balance or /status."
		case "test":
			msg.Text = "I'm ok."
		case "open":
			msg.Text = "What would you like to query?"
			msg.ReplyMarkup = mainKeyboard
		case "close":
			msg.Text = "keyboard closed. Type /open to reopen"
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		case "synced":
			command, _ := botExecCmdOut("celocli node:synced", msg)
			msg.Text = string(command)
		case "balance":
			var valGr validatorGr
			var val validator
			UpdateBalance(&valGr, msg)
			UpdateBalance(&val, msg)
			msgPiece1 := `*gold*: ` + valGr.balance.gold + "\n" + `*lockedGold*: ` + valGr.balance.lockedGold + "\n" + `*usd*: ` + valGr.balance.usd + "\n" + `*non-voting*: ` + valGr.balance.nonVoting + "\n" + `*total*: ` + valGr.balance.total + "\n"
			msgPiece2 := `*gold*: ` + val.balance.gold + "\n" + `*lockedGold*: ` + val.balance.lockedGold + "\n" + `*usd*: ` + val.balance.usd + "\n" + `*non-voting*: ` + val.balance.nonVoting + "\n" + `*total*: ` + val.balance.total + "\n"
			msg.Text = "Validator Group\n\n" + msgPiece1 + "--------------\n" + "Validator\n\n" + msgPiece2
		case "status":
			command, _ := botExecCmdOut("celocli validator:status --validator $CELO_VALIDATOR_ADDRESS", msg)
			words := cmd.ParseCmdOutput(command, "string", "(true|false)\\s*(true|false)\\s*(\\d*)\\s*(\\d*.)", 0)
			wordsSplit := strings.Fields(fmt.Sprintf("%v", words))
			ifElected := wordsSplit[0] + "\n"
			ifFrontRunner := wordsSplit[1] + "\n"
			numProposed := wordsSplit[2] + "\n"
			perctSigned := wordsSplit[3] + "\n"
			message := `*Elected*: ` + ifElected + `*Frontrunner*: ` + ifFrontRunner + `*Proposed*: ` + numProposed + `*Signatures*: ` + perctSigned
			msg.Text = message
		case "score":
			command, _ := botExecCmdOut("celocli validator:show $CELO_VALIDATOR_ADDRESS", msg)
			words := cmd.ParseCmdOutput(command, "string", "score: (\\d.\\d*)", 1)
			msg.Text = `*Score: *` + fmt.Sprintf("%v", words)
		case "lockgold":
			// update balance before locking
			var valGr validatorGr
			var val validator
			UpdateBalance(&valGr, msg)
			UpdateBalance(&val, msg)
			msgPiece1 := boldText("Gold Available\n") + "Validator Group: " + valGr.balance.gold + "\n"
			msgPiece2 := "Validator: " + val.balance.gold + "\n"
			msgPiece3 := "\nHow much would you like to lock?"
			msg.Text = msgPiece1 + msgPiece2 + msgPiece3
			msg.ReplyMarkup = lockGoldKeyboard
		case "exchange":
			var valGr validatorGr
			var val validator
			UpdateBalance(&valGr, msg)
			UpdateBalance(&val, msg)
			msgPiece1 := boldText("USD Available\n") + "Validator Group: " + valGr.balance.usd + "\n"
			msgPiece2 := "Validator: " + val.balance.usd + "\n"
			msgPiece3 := "\nHow much would you like to exchange?\n"
			msg.Text = msgPiece1 + msgPiece2 + msgPiece3
			msg.ReplyMarkup = exchangeUsdKeyboard
		case "vote":
			var valGr validatorGr
			var val validator
			UpdateBalance(&valGr, msg)
			UpdateBalance(&val, msg)
			if valGr.balance.nonVoting == "" && val.balance.nonVoting == "" {
				msg.Text = "You have no non-voting lockedGold available"
			} else {
				msgPiece1 := boldText("Non-voting Locked Gold Available\n") + "Validator Group: " + valGr.balance.nonVoting + "\n"
				msgPiece2 := "Validator: " + val.balance.nonVoting + "\n"
				msgPiece3 := "\nHow much would you like to cast?\n"
				msg.Text = msgPiece1 + msgPiece2 + msgPiece3
				msg.ReplyMarkup = electionVoteKeyboard
			}
		case "signing":
			_, output := botExecCmdOut("celocli validator:signed-blocks --signer $CELO_VALIDATOR_SIGNER_ADDRESS", msg)
			msg.Text = output
		case "exchange_rate":
			msg.Text = getExchangeRate(msg)
		default:
			msg.Text = "Command not yet supported"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

// botExecCmdOut executes commands and returns command outputs
func botExecCmdOut(cmd string, msg tgbotapi.MessageConfig) ([]byte, string) {
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		msg.Text = fmt.Sprint(err) + ": " + string(output)
	} else {
		if string(output) != "" {
			msg.Text = string(output)
		}
	}
	return output, msg.Text
}
