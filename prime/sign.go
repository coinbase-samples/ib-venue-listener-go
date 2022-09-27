package prime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func Sign(channel, key, acctId, t, portfolioId, productIds, signingKey string) string {
	h := hmac.New(sha256.New, []byte(signingKey))
	prodIds := strings.Replace(productIds, "[", "", 1)
	prodIds = strings.Replace(prodIds, "]", "", 1)
	prodIds = strings.Replace(prodIds, "\"", "", -1)
	prodIds = strings.Replace(prodIds, ",", "", -1)
	prodIds = strings.Replace(prodIds, " ", "", -1)

	// Portfolio is blank for prices
	// "#{channelName}#{accessKey}#{svcAcctId}#{timestamp}#{portfolioId}#{productIds}";
	h.Write([]byte(
		fmt.Sprintf(
			"%s%s%s%s%s%s", channel, key, acctId, t, portfolioId, prodIds),
	),
	)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
