from requests import get, post
import argparse
from random import randint, choices
from string import ascii_uppercase, ascii_lowercase, digits
from enum import Enum
from threading import Thread
from time import sleep
from uuid import uuid4
from urllib.request import urlopen


# really just for testing dont make it stupid long
_MAX_PASSWD_LENGTH = 40
_MAX_SLEEP_TIME = 6  # 1 second more than how long we should wait
_SLEEP_CHOICES = [min(x+1, _MAX_SLEEP_TIME) for x in range(_MAX_SLEEP_TIME)]
_MAX_ATTEMPTS = _MAX_SLEEP_TIME  # assume 1 per second

class Actions(Enum):
    # no 0-value
    SHUTDOWN = 1            # 001
    HASH = 2                # 010
    HASH_SHUTDOWN = 3       # 011
    STATS = 4               # 100
    STATS_SHUTDOWN = 5      # 101
    HASH_STATS = 6          # 110
    HASH_STATS_SHUTDOWN = 7 # 111


class RequestTracker:

    def __init__(self):
        # just an example once these are set they shouldn't be changed. 
        # not going to use `property` decorator to discourage that, because
        # this is not for public use
        self.uuid = str(uuid4())
    
        self.request_id = None  # don't override ID what a mistake that could be
        self.original_password = self._generate_password()
        self.hashed_password = None  # only received from a response
    
    def _generate_password(self):
        # generate a password between 20% and 100% of the max length
        this_pwd_len = randint(int(_MAX_PASSWD_LENGTH * 0.2), _MAX_PASSWD_LENGTH)
        return ''.join(choices(ascii_uppercase + ascii_lowercase + digits, k=this_pwd_len))
    

def executeClient(action, lookup):
    """
    Thread that should be executed as an action 
    """
    tracker = RequestTracker()
    
    if "HASH" in action.name:
        print("Tracker {} running HASH".format(tracker.uuid))
        # post the hash request

        _url = lookup.get(Actions.HASH.value, "")
        
        try:
            res = post(_url, data={"password": tracker.original_password})
            
            try:
                tracker.request_id = int(res.text)
                print("Tracker {} received ID {}".format(tracker.uuid, tracker.request_id))            
            except ValueError:
                return                
                
            # wait some number of seconds and try to grab the data that was created in the "Server"
            attempts = 0

            hashed_url = "{}/{}".format(_url, tracker.request_id)

            while not tracker.hashed_password and attempts < _MAX_ATTEMPTS:
                _sleep_time = choices(_SLEEP_CHOICES)[0]
                    
                print("Tracker {} Sleeping {} seconds".format(tracker.uuid, _sleep_time))
                sleep(_sleep_time)
                
                res = post(hashed_url)
            
                if res.text and not "ERROR" in res.text:
                    tracker.hashed_password = res.text
                else:
                    print("Tracker {} trying again".format(tracker.uuid))
                
                attempts = attempts + 1

                
            if not tracker.hashed_password or attempts >= _MAX_ATTEMPTS:
                print("ERROR Tracker {} failed to determine hashed password".format(tracker.uuid))
            else:
                print("Tracker {} Request {}, {} -> {}".format(tracker.uuid, tracker.request_id, tracker.original_password, tracker.hashed_password))
            
        except Exception as e:  # bit wide, but again just testing 
            print(e)
    
    if "STATS" in action.name:
        print("Tracker {} running STATS".format(tracker.uuid))
        _url = lookup.get(Actions.STATS.value, "")
        
        try:
            res = get(_url)
            print("Tracker {} Received STATS data: {}".format(tracker.uuid, res.text))
        except Exception:  
            pass
    
    # TODO: works with curl but not here ...
    if "SHUTDOWN" in action.name:
        print("Tracker {} running SHUTDOWN".format(tracker.uuid))
        _url = lookup.get(Actions.SHUTDOWN.value, "")
        
        try:
            urlopen(_url)
        except:
            print("Tracker {} shut it all down".format(tracker.uuid))


def main(*args, **kwargs):
    """
    Cliche and this may be a small script but lets go ahead and set
    up the main function. 
    
    The purpose of this script will be to test some of the PUT and GET 
    requests sent to the simple Go(lang) http server that was setup.
    """
    url = "http://{}:{}/".format(kwargs.get("host", "localhost"), kwargs.get("port", 8080))
    
    _hash_url = url + "hash"
    _stats_url = url + "stats"
    _shutdown_url = url + "shutdown"
    
    clients = kwargs.get("num_clients", 1)
    
    action_weights = [
        1,  # SHUTDOWN
        40, # HASH
        10, # HASH_SHUTDOWN
        20, # STATS
        1,  # STATS_SHUTDOWN        
        27, # HASH_STATS
        1,  # HASH_STATS_SHUTDOWN
    ]
        
    client_action = choices([x for x in Actions], weights=action_weights, k=clients)
    
    lookup = {
        Actions.HASH.value: _hash_url,
        Actions.STATS.value: _stats_url,
        Actions.SHUTDOWN.value: _shutdown_url
    }
    
    
    threads = [Thread(target=executeClient, args=(client_action[i],lookup)) for i in range(clients)]
    
    for thread in threads:
        thread.start()
        
    
    for thread in threads:
        thread.join()


if __name__ == "__main__":
    
    parser = argparse.ArgumentParser()
    parser.add_argument('host', help='Ip address of the host', type=str)
    parser.add_argument('port', type=int, help='port to connect to')
    parser.add_argument(
        '-n', '--num_clients', type=int, default=1,
        help='number of clients to run and potential number of simulataneous requests'
    )
    
    args=parser.parse_args()
    
    main(**vars(args))
    