package main

import (
    "fmt"
    "os"
    "strconv"
    "net/http"
    "go_web_server/state"
    "crypto/sha512"
    "encoding/base64"
    "time"
    "context"
    "strings"
)

/**
 * Setup a `fake` endpoint for the user by associating a hash to the id. The function
 * will wait `sleep_time_secs` before generating the Sha512 hash of
 * the `passwrd`. After completing the action, the data can be received
 * via a curl request.
 * @param passwd - String to be hashed
 * @param request_id - The interger request that is associated with the passwd
 * @param sleep_time_secs - Number of seconds to `sleep` or wait before valid endpoint
 *
 * Note: Sleep is not the most accurate function, look into making this more
 * at the actual time limit.
 */
func associateHash(passwd string, request_id int, sleep_time_secs int) {
    if len(passwd) > 0 {
       
       // instructions said that the hash SHOULD NOT be computed for 5 seconds
       // technically it doesn't matter when it is computed, just don't do anything
       // with it for 5 seconds, but let's follow instructions
       time.Sleep(5 * time.Second)

       hasher := sha512.New()
       hasher.Write([]byte(passwd))

       // assuming URL encoding due to the nature of the tasking
       // Can just as easily make this StdEncoding
       url_encoded_str := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

       // add the map entry here
       sm := state.GetStateManager()
       sm.AddHash(strconv.Itoa(request_id), url_encoded_str)
    }
}

/**
 * Process the Requests tied to the subrequests for /hash on /hash/xxxxxx.
 * This isn't a true Handle in the sense that a handle was setup specifically
 * for /hash/1. This handler will attempt to find the id and the matching hash
 * returning the correct data to the user. The original design allowed for dynamic
 * creation of Handles, but the shutdown changed the design. Due to the inability to
 * use non built-in modules I decided to make this change. Tehcnically it will scale
 * as it is controlled by a map. 
 */
func processHashedResponse(w http.ResponseWriter, req *http.Request) {

     sm := state.GetStateManager()

     trimmed := strings.ReplaceAll(req.URL.Path, "/hash/", "")

     if hashed_value, isOk := sm.GetHash(trimmed); isOk {
     	fmt.Fprintf(w, hashed_value)
     } else {
       fmt.Fprintf(w, "ERROR, invalid link %s", req.URL.Path)
     }

}

/**
 * Parse data from the /hash endpoint
 * @param w - ResponseWriter where data can be written back to the user.
 *            On error, ERROR is written to the caller. On success the
 *            caller is provided with the ID of the request. This ID can
 *            later be used to lookup the hash of the data that was accepted here.
 * @param req - Original Request information
 */
func hashEndpoint(w http.ResponseWriter, req *http.Request) {

    // inital time of request
    start := time.Now().Nanosecond() / 1000
     
    passwd := ""

    switch req.Method {
    // technically not supporting GET Methods
    case "GET":
	 for k, v := range req.URL.Query() {
	     fmt.Printf("%s: %s\n", k, v)
    	 }
    case "POST":
    	 if err := req.ParseForm(); err != nil {
	    fmt.Printf("ERROR parsing form: %s", err)
    	 } else {
	    passwd = req.Form.Get("password")
	 }
    }

    // end time of request - processed the password, not going to include the time to REPOST
    
    // Technically the the task states that we need to provide the average number of microseconds
    // to process the requests to the /hash endpoint. It does NOT state that we need to process
    // the average time to /hash and its sub-endpoints
    end := time.Now().Nanosecond() / 1000

    // only write valid handles back
    if len(passwd) > 0 {
       // write the value back to the user that sent the request
       sm := state.GetStateManager()

       updated_hash_request := sm.CreateRequest(int(end-start))

       fmt.Fprintf(w, "%d", updated_hash_request)

       // first use of a true go routine
       // Necessary with this design because the write or Fprintf
       // was waiting until the end of the function before we would
       // received it at the client, causing ~5 delay just to get the request id
       go associateHash(passwd, updated_hash_request, 5)
       
    } else {
      fmt.Fprintf(w, "Error: no password provided")
    }
}

/**
 * Parse data from the /stats endpoint
 * @param w - ResponseWriter where data can be written back to the user
 *            In the event of an error, ERROR is written back. On success json
 *            formatted data is written to the caller.
 * @param req - Original Request information
 */
func statsEndpoint(w http.ResponseWriter, req *http.Request) {

     sm := state.GetStateManager()
     total, avg_time_us := sm.GetStatistics()

     json_str := state.CreateResponse(total, avg_time_us)

     if len(json_str) > 0 { 
     	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, json_str)
     } else {
       fmt.Fprint(w, "Error: Failed to produce statistics")
     }
}

/*
 * @param potential - String input that is supposed to be an integer
 * @return True when the potential int can be an integer, False otherwise
*/
func isValidInt(potential string) (bool) {
    _, err := strconv.Atoi(potential)
    return err == nil
}

/**
 * Error Code: 
 *  1) Bad port value
 *  2) No port provided
 *
*/
func main() {
    // grab all args except for the executable name
    server_args := os.Args[1:]

    port := ""

    // parse arguments -- probably could make this more in depth but this currently
    // adds a bit of ease for expansion -- look at making into a function ?
    switch len(server_args) {
    case 1:
        if isValidInt(server_args[0]) {
            port = server_args[0]
        } else {
            os.Exit(1)
        }
    case 0:
        os.Exit(2)
    default:
        for i := 0; i < len(server_args); i++ {
            if (server_args[i] == "-p" || server_args[i] == "-port") && (i+1 < len(server_args)) {
                if isValidInt(server_args[i+1]) {
                    port = server_args[i+1]
                } else {
                    os.Exit(1)
                }
                break  // only here because we have found the param we are looking for
            }
            // unhandled args
        }

        if len(port) == 0 {
            os.Exit(2)
        }
    }

    mux := http.NewServeMux()
    serv := http.Server{Addr: ":"+port, Handler: mux}

    // Setup endpoints for this server
    mux.HandleFunc("/hash", hashEndpoint)
    mux.HandleFunc("/hash/", processHashedResponse)
    mux.HandleFunc("/stats", statsEndpoint)

    // utilize the built-in graceful shutdown
    mux.HandleFunc("/shutdown", func(w http.ResponseWriter, req *http.Request) {
    				       w.Write([]byte("Shutdown Complete"))
    				       if err := serv.Shutdown(context.Background()); err != nil {
				       	  fmt.Println("ERROR: Server graceful shutdown failed")
				       }
    })

    if err := serv.ListenAndServe(); err != nil {
       fmt.Println(err)
    }
}