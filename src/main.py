import json
import requests
import sys


def get_all_clients():  # Gets a list of all known clients
    url = f"{BASE_URL}/clients"
    response = requests.request("GET",url, headers=HEADERS)
    response.raise_for_status()
    return response.json()

def name_matches_prefix(name):  # Checks to see if the comment section of a client starts with a given string
    return any(name.get("comment").startswith(prefix) for prefix in TARGET_CLIENT_PREFIXES)

def reset_clients_with_prefixes():  # main function that performs the client sorting and sends the update request
    clients_json = get_all_clients()  # retrieves list of all known clients and stores it

    clients_list = clients_json.get("clients")
    client_id_list = []
    for item in clients_list:  # iterate over all clients and store them in a separate list if the comment starts with a certain prefix
        client_id_list.append(item) if name_matches_prefix(item) else print(f"Skipping {item.get("comment")}")


    for client in client_id_list:  # update client groups of every client that got moved into the second list
        response = requests.request("PUT", f"{BASE_URL}/clients/{client.get("client")}", headers=HEADERS,
                                    json={"client": f"{client.get("client")}", "name": f"{client.get("name")}",
                                    "comment": f"{client.get("comment")}", "groups": TARGET_GROUPS})
        response.raise_for_status()

    print("Clients with matching prefixes have been reset.")
    return

def get_auth():  # Authenticates with Pi-Hole and gets session ID
    response = requests.request("POST", f"{BASE_URL}/auth", json={"password": f"{API_TOKEN}"})
    response.raise_for_status()
    return response.json().get("session").get("sid")

def show_groups():  # Only used during testing
    response = requests.request("GET", f"{BASE_URL}/groups", headers=HEADERS, json={"password": f"{API_TOKEN}"})
    response.raise_for_status()
    group_info_list = []
    for x in response.json().get("groups"):
        group_info_list.append(f"{x.get("name")} , {x.get("id")}")
    return group_info_list


if __name__ == "__main__":

    with open("config.json", "r") as f:
        env_file = json.load(f)
    f.close()

    BASE_URL = env_file.get("BASE_URL")  # URL to add api endpoints to
    API_TOKEN = env_file.get("API_TOKEN")  # Pi-Hole app password
    TARGET_CLIENT_PREFIXES = env_file.get("TARGET_CLIENT_PREFIXES")  # Client prefixes to search for
    # Gets either group prefixes or group IDs
    TARGET_GROUP_PREFIXES = env_file.get("TARGET_GROUP_PREFIXES") if env_file.get("GROUP_ID_MODE") else TARGET_GROUPS = env_file.get("TARGET_GROUPS")

    sid = get_auth()
    HEADERS = {"X-FTL-SID": f"{sid}", "Accept": "application/json", "Content-Type": "application/json"}

    # uncomment to show group IDs in console
    # print(show_groups())
    # sys.exit(0)

    #reset_clients_with_prefixes()

