package whatsappdau

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Whatsapp interface {
	SendMessage(to string, message string) (*MessageResponse, error)
	SendAudioToWhatsApp(recipientWAID string, filePath string) (string, error)
	SendImageToWhatsApp(recipientWAID string, filePath string) (string, error)
	SendInteractiveList(recipientPhoneNumber string, bodyText string, buttonTitle string, items []ListItem) (*MessageResponse, error)
	SendWhatsAppLocation(recipientPhone string, latitude, longitude float64, name, address string) (*MessageResponse, error)
	SendInteractiveButtons(recipientPhoneNumber string, menuType, bodyText string, buttons []ButtonItem) (*MessageResponse, error)

	GetMediaURL(mediaID string) (*MediaUrl, error)
	DownloadMedia(mediaUrl string) ([]byte, error)
}

type WhatsappClient struct {
	Ctx         context.Context
	apiURL      string
	accessToken string
	client      *http.Client
}

func NewWhatsappClient(ctx context.Context, apiURL string, accessToken string, client *http.Client) Whatsapp {
	log.Printf("apiURL: %s", apiURL)
	log.Printf("accessToken: %s", accessToken)
	return &WhatsappClient{
		Ctx:         ctx,
		apiURL:      apiURL,
		accessToken: accessToken,
		client:      client,
	}
}

func (w *WhatsappClient) SendMessage(recipientWAID string, messageBody string) (*MessageResponse, error) {
	messageData := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipientWAID,
		"type":              "text",
		"text": map[string]string{
			"body": messageBody,
		},
	}

	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	token := fmt.Sprintf("Bearer %s", w.accessToken)
	log.Printf("token: %s", token)
	req.Header.Set("Authorization", token)

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	fmt.Printf("WhatsApp API Response Status: %d\n", resp.StatusCode)
	fmt.Printf("WhatsApp API Response Body: %s\n", string(responseBody))
	var messageResponse MessageResponse
	err = json.Unmarshal(responseBody, &messageResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("failed to send message, status code: %d, response: %s", resp.StatusCode, string(responseBody))
	}

	return &messageResponse, nil
}

func (w *WhatsappClient) SendInteractiveList(recipientPhoneNumber string, bodyText string, buttonTitle string, items []ListItem) (*MessageResponse, error) {
	sections := []ListSection{
		{
			Rows: items,
		},
	}

	interactive := ListInteractive{
		Type: "list",
		Body: BodyText{
			Text: bodyText,
		},
		Action: ListAction{
			Button:   buttonTitle,
			Sections: sections,
		},
	}

	message := WhatsAppMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               recipientPhoneNumber,
		Type:             "interactive",
		Interactive:      interactive,
	}

	return w.sendListMessage(message)
}

func (w *WhatsappClient) SendInteractiveButtons(recipientPhoneNumber string, menuType, bodyText string, buttons []ButtonItem) (*MessageResponse, error) {
	action := ButtonAction{}
	if menuType == "text" {
		return w.SendMessage(recipientPhoneNumber, bodyText)
	}

	if menuType == "location_request_message" {
		action.Name = "send_location"
	}
	for _, btn := range buttons {
		if btn.Link != "" {

			action.Name = "cta_url"
			action.Parameters = &Parameters{
				DisplayText: btn.Text,
				Url:         btn.Link,
			}
		} else {
			backBtn := struct {
				Type  string      `json:"type"`
				Reply ButtonReply `json:"reply"`
			}{
				Type: btn.Type,
				Reply: ButtonReply{
					ID:    btn.ID,
					Title: btn.Text,
				},
			}

			action.Buttons = append(action.Buttons, backBtn)
		}
	}

	interactive := ButtonsInteractive{
		Type: menuType,
		Body: BodyText{
			Text: bodyText,
		},
		Action: action,
	}

	message := WhatsAppMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               recipientPhoneNumber,
		Type:             "interactive",
		Interactive:      interactive,
	}

	return w.sendListMessage(message)
}

func (w *WhatsappClient) sendListMessage(message WhatsAppMessage) (*MessageResponse, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Ошибка кодирования JSON:", err)

	}

	log.Printf("JSON-сообщение: %s", string(jsonData))

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка создания HTTP-запроса:", err)

	}

	req.Header.Set("Content-Type", "application/json")
	token := fmt.Sprintf("Bearer %s", w.accessToken)
	log.Printf("token: %s", token)
	req.Header.Set("Authorization", token)

	resp, err := w.client.Do(req)
	if err != nil {
		fmt.Println("Ошибка отправки HTTP-запроса:", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("Статус код:", resp.Status)

	var response MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Ошибка декодирования JSON ответа:", err)

	}
	fmt.Println("Тело ответа:", response)
	return &response, nil
}

func (w *WhatsappClient) SendAudioToWhatsApp(recipientWAID string, filePath string) (string, error) {
	mediaId, err := w.uploadMedia(filePath, "audio/ogg")
	if err != nil {
		return "", err
	}

	_, err = w.sendWhatsAppMedia(recipientWAID, mediaId)
	if err != nil {
		return "", err
	}
	return mediaId, nil
}

func (w *WhatsappClient) SendImageToWhatsApp(recipientWAID string, filePath string) (string, error) {
	mediaId, err := w.uploadMedia(filePath, "image/jpeg")
	if err != nil {
		return "", err
	}

	err = w.sendWhatsAppImage(recipientWAID, mediaId)
	if err != nil {
		return "", err
	}
	return mediaId, nil
}

func (w *WhatsappClient) uploadMedia(filePath, mediaType string) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file part
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	// Add required fields
	_ = writer.WriteField("type", mediaType)
	_ = writer.WriteField("messaging_product", "whatsapp")

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", w.apiURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, respBody)
	}

	var response struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return response.ID, nil
}

func (w *WhatsappClient) sendWhatsAppMedia(recipientPhone, mediaID string) (*MessageResponse, error) {

	message := AudioMessage{
		MessagingProduct: "whatsapp",
		To:               recipientPhone,
		Type:             "audio",
	}
	message.Audio.ID = mediaID

	body, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	token := fmt.Sprintf("Bearer %s", w.accessToken)
	log.Printf("token: %s", token)
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}
	var response MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Ошибка декодирования JSON ответа:", err)
		return nil, err
	}
	fmt.Println("Audio sent successfully!")
	return &response, nil
}

func (w *WhatsappClient) sendWhatsAppImage(recipientPhone, mediaID string) error {
	message := ImageMessage{
		MessagingProduct: "whatsapp",
		To:               recipientPhone,
		Type:             "image",
	}
	message.Image.ID = mediaID

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	token := fmt.Sprintf("Bearer %s", w.accessToken)
	log.Printf("token: %s", token)
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error: received status code %d - %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Println("Image sent successfully!")
	return nil
}

func (w *WhatsappClient) SendWhatsAppLocation(recipientPhone string, latitude, longitude float64, name, address string) (*MessageResponse, error) {
	message := LocationMessage{
		MessagingProduct: "whatsapp",
		To:               recipientPhone,
		Type:             "location",
	}
	message.Location.Latitude = latitude
	message.Location.Longitude = longitude
	message.Location.Name = name
	message.Location.Address = address

	body, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	token := fmt.Sprintf("Bearer %s", w.accessToken)
	log.Printf("token: %s", token)
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error: received status code %d - %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Println("Location sent successfully!")

	var response MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Ошибка декодирования JSON ответа:", err)
		return nil, err
	}

	return &response, nil
}
func (w *WhatsappClient) GetMediaURL(mediaID string) (*MediaUrl, error) {
	var mediaUrl MediaUrl
	url := fmt.Sprintf("%s/%s", "https://graph.facebook.com/v17.0", mediaID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.accessToken))

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&mediaUrl)
	if err != nil {
		return nil, err
	}

	return &mediaUrl, nil
}

func (w *WhatsappClient) DownloadMedia(mediaUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", mediaUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.accessToken))

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
