package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	c := NewInfraiClient()
	if c.Key == "" {
		log.Fatal("set INFRAI_API_KEY")
	}
	if err := c.CreateBucket(bucket); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		var e PaymentEvent
		if json.NewDecoder(r.Body).Decode(&e) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		d, err := PrepareUpload(c, e)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
