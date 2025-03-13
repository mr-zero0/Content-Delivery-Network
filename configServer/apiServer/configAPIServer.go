package apiServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/hcl/cdn/common/config"
	"github.com/hcl/cdn/configServer/cacheCommander"
	"github.com/hcl/cdn/configServer/configPusher"
	"github.com/hcl/cdn/configServer/configSaver"

	//"github.com/hcl/cdn/configServer/inMemoryConfig"
	"github.com/gorilla/mux" // Use Gorilla Mux for routing
)

func RegisterHandlers(r *mux.Router) {
	// Register routes
	slog.Info("Added /ds")
	r.HandleFunc("/ds", handleDeliveryServicesGet).Methods("GET")
	r.HandleFunc("/ds", handleDeliveryServicesPost).Methods("POST")
	r.HandleFunc("/ds/{name}", handleDeliveryServiceByNameGet).Methods("GET")
	r.HandleFunc("/ds/{name}", handleDeliveryServiceByNamePut).Methods("PUT")
	r.HandleFunc("/ds/{name}", handleDeliveryServiceByNameDelete).Methods("DELETE")
	slog.Info("Added /cn")
	r.HandleFunc("/cn", handleCacheNodesGet).Methods("GET")
	r.HandleFunc("/cn", handleCacheNodesPost).Methods("POST")
	r.HandleFunc("/cn/{name}", handleCacheNodeByNameGet).Methods("GET")
	r.HandleFunc("/cn/{name}", handleCacheNodeByNamePut).Methods("PUT")
	r.HandleFunc("/cn/{name}", handleCacheNodeByNameDelete).Methods("DELETE")
	slog.Info("Added /invalidate")
	r.HandleFunc("/invalidate/{pattern}", handleInvalidate).Methods("GET")
	slog.Info("Added /invalidateStatus")
	r.HandleFunc("/invalidateStatus/{uid}", handleInvalidateStatus).Methods("GET")
}

// --- Delivery Services ---
func handleDeliveryServicesGet(w http.ResponseWriter, r *http.Request) {
	if inMemConfig == nil {
		http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
		return
	}
	response := struct {
		Version     int      `json:"version"`
		ServiceList []string `json:"serviceList"`
	}{}
	response.Version, response.ServiceList = inMemConfig.GetDsNames()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDeliveryServicesPost(w http.ResponseWriter, r *http.Request) {
	//DeSerialize Json
	var newService config.DeliveryService
	if err := json.NewDecoder(r.Body).Decode(&newService); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// Validate that all required fields are present
	if newService.Name == "" || newService.ClientURL == "" || newService.OriginURL == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	for _, rule := range newService.RewriteRules {
		if rule.HeaderName == "" || rule.Operation < 0 || rule.Operation > 2 {
			http.Error(w, "Invalid rewrite rule", http.StatusBadRequest)
			return
		}
	}
	if inMemConfig == nil {
		http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
		return
	}

	_, err := inMemConfig.GetDsDetail(newService.Name)
	if err == nil {
		http.Error(w, "Delivery service already exists", http.StatusConflict)
		return
	}

	err = inMemConfig.AddDs(&newService)
	if err != nil {
		http.Error(w, "Internal error: Error adding ds", http.StatusInternalServerError)
		return
	}
	configSaver.SaveDSToFile()
	configPusher.PushDsUpdate(bgContext)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Delivery service added"))
}
func handleDeliveryServiceByNameGet(w http.ResponseWriter, r *http.Request) {
	if inMemConfig == nil {
		http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	name := vars["name"]
	slog.Info("Handling GET request for service", "ds", name)
	resp, err := inMemConfig.GetDsDetail(name)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func handleDeliveryServiceByNamePut(w http.ResponseWriter, r *http.Request) {
	if inMemConfig == nil {
		http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	name := vars["name"]
	slog.Info("Handling PUT request for service", "ds", name)
	var updatedService config.DeliveryService
	if err := json.NewDecoder(r.Body).Decode(&updatedService); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if updatedService.Name != name {
		http.Error(w, "Name in the URL does not match the name in the body", http.StatusBadRequest)
		return
	}
	err := inMemConfig.UpdateDs(&updatedService)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	configSaver.SaveDSToFile()
	configPusher.PushDsUpdate(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Delivery service updated"))
}
func handleDeliveryServiceByNameDelete(w http.ResponseWriter, r *http.Request) {
	if inMemConfig == nil {
		http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	name := vars["name"]
	slog.Info("Handling DELETE request for service", "ds", name)
	err := inMemConfig.DeleteDs(name)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	configSaver.SaveDSToFile()
	configPusher.PushDsUpdate(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Delivery service deleted"))
}

func handleCacheNodesGet(w http.ResponseWriter, r *http.Request) {
    if inMemConfig == nil {
        http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
        return
    }
    response := struct {
        Version      int      `json:"version"`
		NodeList []string `json:"NodeList"`
        //CacheNodeNames []string `json:"cacheNodeNames"`
    }{}
   // response.Version, response.CacheNodeNames = inMemConfig.GetCnNames()
    response.Version, response.NodeList = inMemConfig.GetCnNames()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
func handleCacheNodesPost(w http.ResponseWriter, r *http.Request) {
    // De-serialize JSON
    var newNode config.CacheNode
    defer r.Body.Close()

    reqBody, err := io.ReadAll(r.Body)
    if err != nil {
        slog.Error("CN Post : Error reading request body", "error", err)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    slog.Info("CN Post : Request", "body", string(reqBody))
    if err := json.NewDecoder(bytes.NewReader(reqBody)).Decode(&newNode); err != nil {
        slog.Error("CN Post : Error decoding req", "error", err)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    // Validate that all required fields are present
    if newNode.Name == "" || newNode.IP == "" || newNode.Port == 0 || newNode.Type == ""  {
        slog.Error("CN Post : Missing required fields", "data", newNode)
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }
    if inMemConfig == nil {
        slog.Error("CN Post : inMemory Config not available")
        http.Error(w, "Internal error: InMemConfig not initialized", http.StatusInternalServerError)
        return
    }
    _, err = inMemConfig.GetCnDetailByName(newNode.Name)
    if err == nil {
        slog.Error("CN Post : CN already exists")
        http.Error(w, "Cache node already exists", http.StatusConflict)
        return
    }
    err = inMemConfig.AddCn(&newNode)
    if err != nil {
        slog.Error("CN Post : CN Add Failed", "error", err)
        http.Error(w, "Internal error: Error adding cn", http.StatusInternalServerError)
        return
    }
    configSaver.SaveCNToFile()
    configPusher.PushCnUpdate(newNode.Name)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("CacheNode service added"))
}

func handleCacheNodeByNameGet(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    slog.Info("Handling request for cache node", "name", name)
    resp, err := inMemConfig.GetCnDetailByName(name)
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func handleCacheNodeByNamePut(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    slog.Info("Handling request for cache node", "name", name)
    var updatedNode config.CacheNode
    if err := json.NewDecoder(r.Body).Decode(&updatedNode); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    if updatedNode.Name != name {
        http.Error(w, "Name in the URL does not match the name in the body", http.StatusBadRequest)
        return
    }
    err := inMemConfig.UpdateCn(&updatedNode)
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    configSaver.SaveCNToFile()
    configPusher.PushCnUpdate(name)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Cache Node updated"))
}

func handleCacheNodeByNameDelete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    slog.Info("Handling request for cache node", "name", name)
    err := inMemConfig.DeleteCn(name)
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    configSaver.SaveCNToFile()
    configPusher.PushCnUpdate(name)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Cache Node deleted"))
}
func generateUniqueID() string {
	timestamp := time.Now().UnixNano()
	randomNumber := rand.Int63()
	return fmt.Sprintf("%d-%d", timestamp, randomNumber)
}

func handleInvalidate(w http.ResponseWriter, r *http.Request) {
	//var inMemConfig *inMemoryConfig.InMemoryConfig
	vars := mux.Vars(r)
	pattern := vars["pattern"]

	invalidationID := generateUniqueID()
	cacheCommander.ExecuteInvalidateRequest(invalidationID, pattern)

	w.Header().Add("Location", "/invalidateStatus/"+invalidationID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Invalidation request received with ID: " + invalidationID))
}

func handleInvalidateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invalidationID := vars["uid"]
	statusString := cacheCommander.GetInvalidateRequestStatus(invalidationID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{ \"status\" : " + statusString + " }"))
}
