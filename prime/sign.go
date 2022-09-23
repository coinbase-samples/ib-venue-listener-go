package prime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func Sign(channel, key, acctId, ts, portfolioId, productIds, signingKey string) string {
	h := hmac.New(sha256.New, []byte(signingKey))
	prodIds := strings.Replace(productIds, "[", "", 1)
	prodIds = strings.Replace(prodIds, "]", "", 1)
	prodIds = strings.Replace(prodIds, "\"", "", -1)
	prodIds = strings.Replace(prodIds, ",", "", -1)
	//portfolio is blank for prices
	//string = "#{channelName}#{accessKey}#{svcAcctId}#{timestamp}#{portfolioId}#{productIds}";
	msg := fmt.Sprintf("%s%s%s%s%s%s", channel, key, acctId, ts, portfolioId, prodIds)
	fmt.Println("signing", msg)
	h.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
