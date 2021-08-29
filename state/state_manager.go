package state

import (
       "sync"
       "encoding/json"
)

/**
 * json response to stats queries
 * This could be done with a map for a small application
 * but why not, this gives us defined structure and easy expansion
 */
type stats_response struct {
     Total              int  `json:"total"`
     AverageProcessTime int  `json:"average"`
}

/**
 * Fill stats_response with the data passed in and
 * create a json string that will be the response
 * @param total - number of entries
 * @param average - time in microseconds on average for each response
 */
func CreateResponse (total int, average int) (string) {
     resp := &stats_response{
     	  Total: total,
	  AverageProcessTime: average}

     resp_marshal, err := json.Marshal(resp)
     if err == nil {
     	return string(resp_marshal)
     } else {
       return ""
     }
}

// at least try to keep the data synchronized
var lock = &sync.Mutex{}

type state_manager struct {
     process_times []int

     id_lookup     map[string] string  // key: request id, value: hashed value
}

// Create a singleton that will track the state of the executable
var instance *state_manager

/**
 * Get the instance of the state_manager singleton. If this is the
 * first time being called, the state_manager will be initialized.
 * @return - state_manager singleton instance
 */
func GetStateManager() *state_manager {

     lock.Lock()
     defer lock.Unlock()

     // only allow the creation once ... Singleton 
     if instance == nil {
     	instance = &state_manager{process_times: nil, id_lookup: make(map[string]string)}
     }
     return instance
}

/**
 * @return - Current number of requests
 */
func (sm *state_manager) GetHashRequest() int {
     // this function probably does Not need the lock
     lock.Lock()
     defer lock.Unlock()
     
     return len(sm.process_times)
}

/**
 * Create a request which happens to be just keeping track of how long it took
 * to complete the request (microseconds). As the number is incrementing based on
 * each request, we can just use the length to grab that value.
 *
 * @param time_to_process - Microseconds it took to process this current request
 * @return - current number of requests
 */
func (sm *state_manager) CreateRequest(time_to_process int) int {
     lock.Lock()
     defer lock.Unlock()

     sm.process_times = append(sm.process_times, time_to_process)

     // Can do this since its an incrementing list
     return len(sm.process_times)
}

/**
 * Determine the average time to complete the requests based on the number of
 * requests currently stored in the Singleton
 *
 * @return Average time to complete the requests
 */
func (sm *state_manager) AvgProcessTime() int {
     lock.Lock()
     defer lock.Unlock()

     total := 0

     for _, t := range sm.process_times {
     	 total += t
     }

     if len(sm.process_times) > 0 {
     	return total / len(sm.process_times)
     } else {
       return 0
     }
}

/**
 * Getter function for the statistics information associated with
 * the state singleton
 * @return - number of entries, Average time to complete request
 */
func (sm *state_manager) GetStatistics() (int, int) {
     return sm.GetHashRequest(), sm.AvgProcessTime()
}

/**
 * Add an id and hashed password to the state manager
 * @param id - key to the map (string version of a request id )
 * @param hash - hashed password
 */
func (sm *state_manager) AddHash(id string, hash string) {
     lock.Lock()
     defer lock.Unlock()

     sm.id_lookup[id] = hash
}

/**
 * Find (if exists) the value associated with the id (key)
 * @param id - key to the map (string version of a request id)
 * @return The value, and a boolean stating whether it was actually found or not
 */
func (sm *state_manager) GetHash(id string) (string, bool) {
     lock.Lock()
     defer lock.Unlock()

     value, found := sm.id_lookup[id]

     return value, found
}