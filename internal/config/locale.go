package config

import (
	"github.com/go-playground/locales/en"
	ja "github.com/go-playground/locales/ja"
	tw "github.com/go-playground/locales/zh_Hant_TW"
	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
	tw_translations "github.com/go-playground/validator/v10/translations/zh_tw"
)

type LocaleTranslators struct {
	// Validator is a configured instance which cached structs
	Validator *v.Validate

	// Translators provides translation capabilities for validation messages
	Translators *ut.UniversalTranslator
}

// NewLocaleTranslations creates a new LocaleTranslators instance with support for
// English, Traditional Chinese (Taiwan), and Japanese languages.
// It configures both validation and translation for these locales.
//
// Usage example:
//
//	// Initialize translations
//	locales := config.NewLocaleTranslations()
//
//	// Validate a struct
//	type User struct {
//	    Name  string `validate:"required"`
//	    Email string `validate:"required,email"`
//	}
//
//	user := User{Name: "", Email: "invalid"}
//	err := locales.Validator.Struct(user)
//
//	// Get translator for user's language
//	if trans, ok := locales.Translators.GetTranslator("ja"); ok {
//	    // Translate validation errors
//	    for _, err := range err.(validator.ValidationErrors) {
//	        fmt.Println(err.Translate(trans))
//	        // Output in Japanese: "名前は必須フィールドです"
//	    }
//	}
//
// Supported languages:
//   - English (en) / fallback lang
//   - Japanese (ja)
//   - Traditional Chinese (zh_Hant_TW)
func NewLocaleTranslations() *LocaleTranslators {
	validator := v.New(v.WithRequiredStructEnabled())
	en := en.New()
	tw := tw.New()
	ja := ja.New()
	translators := ut.New(en, en, ja, tw)
	if trans, ok := translators.GetTranslator("en"); ok {
		en_translations.RegisterDefaultTranslations(validator, trans)
	}
	if trans, ok := translators.GetTranslator("zh_Hant_TW"); ok {
		tw_translations.RegisterDefaultTranslations(validator, trans)
	}
	if trans, ok := translators.GetTranslator("ja"); ok {
		ja_translations.RegisterDefaultTranslations(validator, trans)
	}

	return &LocaleTranslators{
		Validator:   validator,
		Translators: translators,
	}
}
