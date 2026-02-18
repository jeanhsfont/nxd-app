
package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
)

// GenerateVertexReport gera um relatório usando o Vertex AI.
// Retorna o relatório gerado e um erro, se houver.
func GenerateVertexReport(prompt string) (string, error) {
	projectID := os.Getenv("VERTEX_AI_PROJECT")
	if projectID == "" {
		return "", fmt.Errorf("variável de ambiente VERTEX_AI_PROJECT não definida")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, projectID, "us-central1")
	if err != nil {
		log.Printf("Erro ao criar cliente GenAI: %v", err)
		return "", fmt.Errorf("erro ao inicializar a IA: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Erro ao gerar conteúdo com Vertex AI: %v", err)
		return "", fmt.Errorf("erro ao gerar relatório com a IA: %w", err)
	}

	var generatedText strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					generatedText.WriteString(string(txt))
				}
			}
		}
	}

	if generatedText.Len() == 0 {
		return "", fmt.Errorf("a IA não retornou conteúdo")
	}

	return generatedText.String(), nil
}
