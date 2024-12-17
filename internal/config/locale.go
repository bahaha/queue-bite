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
	Validator   *v.Validate
	Translators *ut.UniversalTranslator
}

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
