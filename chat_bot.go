package pachca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ChatBotHandler interface {
	ButtonHandler(ctx context.Context, webhook *WebhookButton) error
	MessageHandler(ctx context.Context, webhook *WebhookMessage) error
}

type ChatBot struct {
	client          *Client
	botUserID       int
	responseTimeout time.Duration
	handler         ChatBotHandler
}

var (
	ErrReadingBody       = fmt.Errorf("reading body error")
	ErrWebhookType       = fmt.Errorf("incorrect webhook type")
	ErrWebhookParsing    = fmt.Errorf("parsing webhook error")
	ErrWebhookProcessing = fmt.Errorf("processing webhook error")
)

func NewChatBot(client *Client, handler ChatBotHandler, botUserID int, responseTimeout time.Duration) *ChatBot {
	return &ChatBot{
		client:          client,
		botUserID:       botUserID,
		responseTimeout: responseTimeout,
		handler:         handler,
	}
}

func (cb *ChatBot) Handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		//logger.Zap.Error(fmt.Sprintf("%v: %v", ErrReadingBody, err))
		http.Error(w, ErrReadingBody.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var webhookType *WebhookType
	err = json.Unmarshal(body, &webhookType)
	if err != nil {
		//logger.Zap.Error(fmt.Sprintf("%v: %v", ErrWebhookType, err))
		http.Error(w, ErrWebhookType.Error(), http.StatusBadRequest)
		return
	}

	switch webhookType.Type {

	case "message":
		var webhook *WebhookMessage
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			//logger.Zap.Error(fmt.Sprintf("%v: %v", ErrWebhookParsing, err))
			http.Error(w, ErrWebhookParsing.Error(), http.StatusBadRequest)
			return
		}

		// не реагируем на собственные сообщения
		if webhook.UserID == cb.botUserID {
			return
		}

		// проверяем что текст сообщения не пустой
		if webhook.Content == "" {
			//logger.Zap.Warn("Empty message content")
			http.Error(w, ErrWebhookParsing.Error(), http.StatusBadRequest)
			return
		}

		// не используем родительский контекст, потому что дальше обработка будет выполняться асинхронно,
		// а для текущего http-запроса будет отдан 200 код ответа
		requestCtx, _ := context.WithTimeout(context.Background(), cb.responseTimeout)

		// отдаём в асинхронную обработку вебхук, а клиенту возвращаем ОК
		go func(ctx context.Context, webhook *WebhookMessage) {
			err := cb.handler.MessageHandler(ctx, webhook)
			//if err != nil {
			//	logger.Zap.Error(fmt.Sprintf("error processing received message: %v", err))
			//}
		}(requestCtx, webhook)

	case "button":
		var webhook *WebhookButton
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			//logger.Zap.Error(fmt.Sprintf("%v: %v", ErrWebhookParsing, err))
			http.Error(w, ErrWebhookParsing.Error(), http.StatusBadRequest)
			return
		}

		// проверяем что в кнопке есть данные
		if webhook.Data == "" {
			//logger.Zap.Warn("Empty message content")
			http.Error(w, ErrWebhookParsing.Error(), http.StatusBadRequest)
			return
		}

		// не используем родительский контекст, потому что дальше обработка будет выполняться асинхронно,
		// а для текущего http-запроса будет отдан 200 код ответа
		requestCtx, _ := context.WithTimeout(context.Background(), cb.responseTimeout)

		// отдаём в асинхронную обработку вебхук, а клиенту возвращаем ОК
		go func(ctx context.Context, webhook *WebhookButton) {
			err := cb.handler.ButtonHandler(ctx, webhook)
			//if err != nil {
			//	logger.Zap.Error(fmt.Sprintf("error processing received message: %v", err))
			//}
		}(requestCtx, webhook)

	default:
		//logger.Zap.Error("This type of the webhook is not implemented")
		http.Error(w, "This type of the webhook is not implemented", http.StatusBadRequest)
	}

	resp := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "OK",
		Message: "Webhook received successfully",
	}
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}
