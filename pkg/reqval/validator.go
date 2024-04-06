package reqval

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/pkg/dupcheck"
	"github.com/PoorMercymain/bannerify/pkg/mimecheck"
)

func ValidateJSONRequest(r *http.Request) error {
	if !mimecheck.IsJSONContentTypeCorrect(r) {
		return appErrors.ErrWrongMIME
	}

	bytesToCheck, err := io.ReadAll(r.Body)
	if err != nil {
		return appErrors.ErrSomethingWentWrong
	}

	reader := bytes.NewReader(bytes.Clone(bytesToCheck))

	err = dupcheck.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		return appErrors.ErrWrongJSON
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bytesToCheck))

	return nil
}
