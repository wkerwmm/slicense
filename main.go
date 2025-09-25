package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"database/sql"

	"license-server/database"
	"license-server/license"
	"license-server/utils"
	"license-server/web"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

func main() {
	utils.LoadConfig("config.yml")
	app := &cli.App{
		Name:  "license-server",
		Usage: "Lisans yönetim sunucusu ve CLI aracı",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Yeni lisans ekle",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "key", Usage: "Lisans anahtarı veya 'random'", Required: true},
					&cli.StringFlag{Name: "product", Usage: "Ürün adı", Required: true},
					&cli.StringFlag{Name: "email", Usage: "Sahibin e-posta adresi", Required: true},
					&cli.StringFlag{Name: "name", Usage: "Sahibin adı", Required: true},
					&cli.IntFlag{Name: "hours", Usage: "Lisans süresi saat cinsinden (opsiyonel)"},
				},
				Action: handleAdd,
			},
			{
				Name:      "delete",
				Usage:     "Lisans sil",
				ArgsUsage: "<key> <product>",
				Action:    handleDelete,
			},
			{
				Name:      "list",
				Usage:     "Lisansları listele",
				ArgsUsage: "<product>",
				Action:    handleList,
			},
			{
				Name:      "logs",
				Usage:     "Audit logları göster",
				ArgsUsage: "[limit]",
				Action:    handleLogs,
			},
			{
				Name:   "serve",
				Usage:  "HTTP sunucusunu başlat",
				Action: handleServe,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getService() (*license.Service, error) {
	cfg := utils.AppConfig

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
	)

	db, err := database.New(dsn)
	if err != nil {
		return nil, err
	}
	return license.NewService(db), nil
}

func getDB() (*sql.DB, error) {
	cfg := utils.AppConfig

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
	)

	return sql.Open("mysql", dsn)
}

func handleAdd(c *cli.Context) error {
	service, err := getService()
	if err != nil {
		return err
	}

	key := c.String("key")
	if key == "random" {
		key = utils.GenerateLicenseKey()
		fmt.Println("Oluşturulan lisans anahtarı:", key)
	}

	product := c.String("product")
	email := c.String("email")
	name := c.String("name")

	var expiresAt *time.Time
	if h := c.Int("hours"); h > 0 {
		exp := time.Now().Add(time.Duration(h) * time.Hour)
		expiresAt = &exp
	}

	err = service.AddLicense(key, product, email, name, expiresAt)
	if err != nil {
		return fmt.Errorf("Lisans eklenemedi: %w", err)
	}

	if expiresAt != nil {
		fmt.Printf("Lisans eklendi: %s (Ürün: %s, Sahip: %s <%s>, Bitiş: %s)\n",
			key, product, name, email, expiresAt.Format(time.RFC822))
	} else {
		fmt.Printf("Süresiz lisans eklendi: %s (Ürün: %s, Sahip: %s <%s>)\n",
			key, product, name, email)
	}

	return nil
}

func handleDelete(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.Exit("Kullanım: delete <key> <product>", 1)
	}

	service, err := getService()
	if err != nil {
		return err
	}

	key := c.Args().Get(0)
	product := c.Args().Get(1)

	err = service.DeleteLicense(key, product)
	if err != nil {
		return fmt.Errorf("Lisans silinemedi: %w", err)
	}

	fmt.Printf("Lisans silindi: %s (Ürün: %s)\n", key, product)
	return nil
}

func handleList(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("Kullanım: list <product>", 1)
	}

	service, err := getService()
	if err != nil {
		return err
	}

	product := c.Args().Get(0)
	licenses, err := service.ListLicenses(product)
	if err != nil {
		return err
	}

	if len(licenses) == 0 {
		fmt.Println("Bu ürün için lisans bulunamadı.")
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Anahtar", "Sahip", "E-posta", "Aktif", "Bitiş Tarihi"})

	for _, lic := range licenses {
		expires := "Süresiz"
		if lic.ExpiresAt != nil {
			expires = lic.ExpiresAt.Format("2006-01-02")
		}
		activated := "Hayır"
		if lic.IsActivated {
			activated = "Evet"
		}
		t.AppendRow(table.Row{lic.Key, lic.OwnerName, lic.OwnerEmail, activated, expires})
	}

	t.Render()
	return nil
}

func handleLogs(c *cli.Context) error {
	service, err := getService()
	if err != nil {
		return err
	}

	limit := 10
	if c.NArg() >= 1 {
		if l, err := strconv.Atoi(c.Args().Get(0)); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := service.GetAuditLogs(limit)
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Tarih", "Aksiyon", "Lisans", "Ürün", "Detaylar"})

	for _, logEntry := range logs {
		t.AppendRow(table.Row{
			logEntry.ChangedAt.Format("2006-01-02 15:04"),
			logEntry.Action,
			logEntry.LicenseKey,
			logEntry.Product,
			logEntry.Details,
		})
	}

	t.Render()
	return nil
}

func handleServe(c *cli.Context) error {
	service, err := getService()
	if err != nil {
		return err
	}

	db, err := getDB()
	if err != nil {
		return err
	}

	licenseHandler := license.NewHandler(service)
	webRouter := web.SetupRoutes(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/license/verify", licenseHandler.VerifyLicense)
	mux.HandleFunc("/license/audit-logs", licenseHandler.GetAuditLogs)
	mux.Handle("/api/", webRouter)

	port := utils.AppConfig.Server.Port
	fmt.Printf("Sunucu http://localhost:%d adresinde çalışıyor\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}