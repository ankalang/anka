package install

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func valid(module string) bool {
	return strings.HasPrefix(module, "github.com/")
}

func Install(module string) {
	if !valid(module) {
		fmt.Printf(`URL formatında hata bulundu. Urller: "github.com/USER/REPO" formatında olmalı.`)
		return
	}

	err := getZip(module)
	if err != nil {
		return
	}

	err = unzip(module)
	if err != nil {
		fmt.Printf(`%v'i paketten çıkarken hata oldu`, err)
		return
	}

	alias, err := createAlias(module)
	if err != nil {
		return
	}

	fmt.Printf("\u001b[32mBaşarı ile indirildi, \u001b[34m`src(\"%s\")` \u001b[32mile kullanabilirsiniz.\u001b[37m\n", alias)
	return
}

func printLoader(done chan int64, message string) {
	var stop = false
	symbols := []string{"🌑 ", "🌒 ", "🌓 ", "🌔 ", "🌕 ", "🌖 ", "🌗 ", "🌘 "}
	i := 0

	for {
		select {
		case <-done:
			stop = true
		default:
			fmt.Printf("\r" + symbols[i] + " - " + message)
			time.Sleep(100 * time.Millisecond)
			i++
			if i > len(symbols)-1 {
				i = 0
			}
		}

		if stop {
			break
		}
	}
}

func getZip(module string) error {
	path := fmt.Sprintf("./paketler/%s-master.zip", module)
	err := os.MkdirAll(filepath.Dir(path), 0755)

	if err != nil {
		fmt.Printf("Klasör oluştunurken hata oluştu %s\n", err)
		return err
	}

	out, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		fmt.Printf("Dosya açılırken hata oluştu %s\n", err)
		return err
	}
	defer out.Close()

	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}

	url := fmt.Sprintf("https://%s/archive/master.zip", module)

	done := make(chan int64)
	go printLoader(done, "Arşiv indiriliyor")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Request gönderilirken hata oluştu %s", err)
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Modül bulunamadı: %s\n", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Errorf("Kötü cevap kodu: %d", resp.StatusCode)
		return err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Dosyalar kopyalanırken hata oluştu %s", err)
		return err
	}
	done <- 1
	return err
}

func unzip(module string) error {
	fmt.Printf("Klasörden çıkarılıyor...\n")
	src := fmt.Sprintf("./paketler/%s-master.zip", module)
	dest := filepath.Dir(src)

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		filename := f.Name
		parts := strings.Split(f.Name, string(os.PathSeparator))
		if len(parts) > 1 {
			if strings.HasSuffix(parts[0], "-master") {
				parts[0] = strings.TrimSuffix(parts[0], "-master")
				filename = strings.Join(parts, string(os.PathSeparator))
			}
		}
		fpath := filepath.Join(dest, filename)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	err = os.Remove(src)
	if err != nil {
		fmt.Printf("%s adresinde zip arşivden çıkarılamadı, hata: %s\n", src, err)
	}

	return nil
}

func createAlias(module string) (string, error) {
	fmt.Printf("Alias açılamadı...\n")
	f, err := os.OpenFile("./paketler.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Alias klasörü ayrıştırılamadı %s\n", err)
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)

	data := make(map[string]string)
	moduleName := filepath.Base(module)
	modulePath := fmt.Sprintf("./paketler/%s", module)

	if len(b) == 0 {
		data[moduleName] = modulePath
	} else {
		err = json.Unmarshal(b, &data)
		if err != nil {
			fmt.Printf("%s\n", err)
			return "", err
		}
		if data[moduleName] == modulePath {
			return moduleName, nil
		}

		if data[moduleName] != "" {
			fmt.Printf("Bu modül, aynı isimde bir modül olduğu için indirilemedi.\n")
			return modulePath, nil
		}

		data[moduleName] = modulePath
	}

	newData, err := json.MarshalIndent(data, "", "    ")

	if err != nil {
		fmt.Printf("Alias json yüklenemedi %s\n", err)
		return "", err
	}

	_, err = f.WriteAt(newData, 0)
	if err != nil {
		fmt.Printf("Alias dosyası değiştirilemedi. %s\n", err)
		return "", err
	}
	return moduleName, err

}
