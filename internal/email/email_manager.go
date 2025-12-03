package email

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"sync"
	textTemplate "text/template"
)

type EmailManager struct {
	EmailTemplateFolder string
	TextCache           map[string]*textTemplate.Template
	HTMLCache           map[string]*htmlTemplate.Template
	mu                  sync.RWMutex
}

func NewEmailManager(templateFolder string) *EmailManager {
	return &EmailManager{
		EmailTemplateFolder: templateFolder,
		TextCache:           make(map[string]*textTemplate.Template),
		HTMLCache:           make(map[string]*htmlTemplate.Template),
	}
}

func (em *EmailManager) getTemplates(templateName string) (string, string, error) {
	folder := filepath.Join(em.EmailTemplateFolder, templateName)
	textPath := filepath.Join(folder, templateName+".txt")
	htmlPath := filepath.Join(folder, templateName+".html")

	textContent, err := os.ReadFile(textPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read text template %s: %w", textPath, err)
	}

	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read HTML template %s: %w", htmlPath, err)
	}

	return string(textContent), string(htmlContent), nil
}

func (em *EmailManager) getOrParseTextTemplate(templateName, content string) (*textTemplate.Template, error) {
	em.mu.RLock()
	if tmpl, exists := em.TextCache[templateName]; exists {
		em.mu.RUnlock()
		return tmpl, nil
	}
	em.mu.RUnlock()

	tmpl, err := textTemplate.New(templateName).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text template: %w", err)
	}

	em.mu.Lock()
	em.TextCache[templateName] = tmpl
	em.mu.Unlock()

	return tmpl, nil
}

func (em *EmailManager) getOrParseHTMLTemplate(templateName, content string) (*htmlTemplate.Template, error) {
	em.mu.RLock()
	if tmpl, exists := em.HTMLCache[templateName]; exists {
		em.mu.RUnlock()
		return tmpl, nil
	}
	em.mu.RUnlock()

	tmpl, err := htmlTemplate.New(templateName).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	em.mu.Lock()
	em.HTMLCache[templateName] = tmpl
	em.mu.Unlock()

	return tmpl, nil
}

func (em *EmailManager) GenerateTemplate(templateName string, properties any) (string, string, error) {
	templateText, templateHTML, err := em.getTemplates(templateName)
	if err != nil {
		return "", "", fmt.Errorf("failed to load templates: %w", err)
	}

	textTmpl, err := em.getOrParseTextTemplate(templateName, templateText)
	if err != nil {
		return "", "", err
	}

	htmlTmpl, err := em.getOrParseHTMLTemplate(templateName, templateHTML)
	if err != nil {
		return "", "", err
	}

	var compiledText bytes.Buffer
	if err := textTmpl.Execute(&compiledText, properties); err != nil {
		return "", "", fmt.Errorf("failed to execute text template: %w", err)
	}

	var compiledHTML bytes.Buffer
	if err := htmlTmpl.Execute(&compiledHTML, properties); err != nil {
		return "", "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return compiledText.String(), compiledHTML.String(), nil
}

func (em *EmailManager) ClearCache() {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.TextCache = make(map[string]*textTemplate.Template)
	em.HTMLCache = make(map[string]*htmlTemplate.Template)
}

func (em *EmailManager) PreloadTemplates(templateNames ...string) error {
	for _, name := range templateNames {
		templateText, templateHTML, err := em.getTemplates(name)
		if err != nil {
			return fmt.Errorf("failed to preload template %s: %w", name, err)
		}

		if _, err := em.getOrParseTextTemplate(name, templateText); err != nil {
			return err
		}

		if _, err := em.getOrParseHTMLTemplate(name, templateHTML); err != nil {
			return err
		}
	}
	return nil
}
