package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func SaveData(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func Utf8ToGbk(s []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return "", e
	}
	return string(d), nil
}

// Output struct as json format
func PrintStruct(s interface{}) string {
	b, err := json.Marshal(s)

	if err != nil {
		return fmt.Sprintf("%+v", s)
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", s)
	}

	return out.String()
}

func GetChineseCharacter(s string) string {
	var result string
	for _, v := range s {
		if v >= 0x4e00 && v <= 0x9fa5 {
			result += string(v)
		}
	}
	return result

	// return regexp.MustCompile("[^\u4e00-\u9fa5]").ReplaceAllString(s, "")
}
