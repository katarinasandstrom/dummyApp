package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"dummyApp/auth"
	reporterv1 "dummyApp/gen/proto/reporter/v1"
	"dummyApp/proto/reporter/v1/reporterv1connect"

	"connectrpc.com/connect"
	// relationv1 "dummyApp/gen/proto/relation/v1"
	// operatorv1 "dummyApp/gen/proto/operator/v1"
)

type client struct {
	reporter reporterv1connect.ReporterServiceClient
	logger   *log.Logger
}

func main() {

	//Steg 1: JWT-anrop
	jwt, err := auth.GetJWT()
	if err != nil {
		log.Fatalf("Failed to get JWT: %v", err)
	}
	fmt.Println("JWT fetched:", jwt)

	// Steg 2: Skapa en klient
	client := NewClient("https://test-backstage.stim.se", log.Default())

	// Steg 3: Hämta listan av reporters
	client.ListAllReporters(context.Background(), jwt)
	if err != nil {
		log.Fatalf("Failed to list reporters: %v", err)
	}

	//  Steg 4: Formatera till JSON
	// jsonData, err := json.MarshalIndent(reporters, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Failed converting to JSON: %v", err)
	// }

	// Steg 5: Skriv ut JSON-formatet

}

func NewClient(host string, logger *log.Logger) *client { //använder en standard-http-client ist för STIMs specialare. Kolla på den sen
	httpclient := &http.Client{}
	return &client{
		reporter: reporterv1connect.NewReporterServiceClient(
			httpclient,
			host,
			connect.WithGRPC(),
		),
		logger: logger,
	}
}

func (c *client) ListReporters(ctx context.Context, jwt string) (*reporterv1.ListResponse, error) {
	req := connect.NewRequest(&reporterv1.ListRequest{
		Limit:  10,
		Offset: 0,
	})

	req.Header().Set("Authorization", "Bearer "+jwt)
	req.Header().Set("Stim-App", "Backstage")

	resp, err := c.reporter.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Could not list reporters: %w", err)
	}

	return resp.Msg, nil
}
func (c *client) ListAllReporters(ctx context.Context, jwt string) {
	var allReporters []*reporterv1.Reporter
	offset := 0
	limit := 50       // Hämta 50 åt gången
	totalFetched := 0 // Räknar antalet hämtade totalt, numrering

	logFile, err := os.OpenFile("reporters.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	//O_CREATE => Skapar filen om den inte finns, O_WRONLY öppnar i WriteOnly,
	//  O_TRUNC skriver över gammalt innehåll
	if err != nil {
		fmt.Errorf("Kunde inte öppna loggfilen: %w", err)
	}

	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	for {
		req := connect.NewRequest(&reporterv1.ListRequest{
			Limit:  int64(limit),
			Offset: int64(offset),
		})

		req.Header().Set("Authorization", "Bearer "+jwt)
		req.Header().Set("Stim-App", "Backstage")

		resp, err := c.reporter.List(ctx, req)
		if err != nil {
			fmt.Errorf("Kunde inte hämta reporters: %w", err)
		}

		for _, r := range resp.Msg.Reporters {
			totalFetched++

			id := "<saknas>"
			if r.Id != nil {
				id = *r.Id
			}

			firstname := "<saknas>"
			if r.Firstname != nil {
				firstname = *r.Firstname
			}

			lastname := "<saknas>"
			if r.Lastname != nil {
				lastname = *r.Lastname
			}

			// Skapa och logga formaterad sträng
			entry := fmt.Sprintf("[%d] ID: %s, Namn: %s %s", totalFetched, id, firstname, lastname)
			logger.Println(entry)
		}

		allReporters = append(allReporters, resp.Msg.Reporters...)

		// Om vi fick färre än "limit", då är vi klara
		if len(resp.Msg.Reporters) < limit {
			break
		}

		offset += limit
	}
}
