package ginbinder

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func HandleBinding(err error) (int, *Response) {
	var (
		verrs validator.ValidationErrors
	)

	if ok := errors.As(err, &verrs); ok {
		var mapErrors string

		for _, verr := range verrs {
			switch verr.Tag() {
			case "required":
				mapErrors += "Поле " + verr.Field() + " обязательно к заполнению\n"
			case "uuid":
				mapErrors += "Поле" + verr.Field() + " должно быть валидным UUID\n"
			case "oneof":
				mapErrors += "Поле" + verr.Field() + " должно принимать одно из следующих значений: " + verr.Param() + "\n"
			case "gte":
				mapErrors += "Поле " + verr.Field() + " должно быть ≥ " + verr.Param() + "\n"
			case "lte":
				mapErrors += "Поле " + verr.Field() + " должно быть ≤ " + verr.Param() + "\n"
			case "hexadecimal":
				mapErrors += "Поле " + verr.Field() + " должно содержать только шестнадцатеричные символы\n"
			case "required_without":
				mapErrors += "Одно из этих полей обязательно к заполнению\n"
			case "len":
				mapErrors += "Поле " + verr.Field() + " должно быть длиной ровно " + verr.Param() + " символов\n"
			case "min":
				mapErrors += "Поле " + verr.Field() + " должно содержать не менее " + verr.Param() + " элементов или символов\n"
			case "max":
				mapErrors += "Поле " + verr.Field() + " должно содержать не более " + verr.Param() + " элементов или символов\n"
			default:
				mapErrors += "Поле " + verr.Field() + " не прошло проверку " + verr.Tag() + "\n"
			}
		}

		return http.StatusBadRequest, &Response{
			Success: false,
			Message: mapErrors,
		}
	}

	return http.StatusBadRequest, &Response{
		Success: false,
		Message: "Переданы неверные параметры запроса",
	}
}
