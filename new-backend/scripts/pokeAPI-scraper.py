import requests
import json
import os
from concurrent.futures import ThreadPoolExecutor, as_completed
import time

# Base URL for PokeAPI
POKEAPI_BASE_URL = "https://pokeapi.co/api/v2/pokemon/"
# Number of concurrent workers (adjust based on your connection and API rate limits)
MAX_CONCURRENT_REQUESTS = 20

def fetch_pokemon_list():
    """
    Fetches the initial list of all Pokémon URLs from PokeAPI.
    """
    print("Fetching list of all Pokémon URLs...")
    url = f"{POKEAPI_BASE_URL}?limit=10000" # Attempt to get all in one request

    try:
        response = requests.get(url, timeout=30) # Add a timeout for the list request
        response.raise_for_status()
        data = response.json()
        
        # Extract only the URLs for detailed fetching
        pokemon_urls = [p['url'] for p in data.get('results', [])]
        print(f"Found {len(pokemon_urls)} Pokémon URLs.\n")
        return pokemon_urls
    except requests.exceptions.RequestException as e:
        print(f"Error fetching Pokémon list: {e}")
        return []

def fetch_pokemon_detail(url):
    """
    Fetches detailed information for a single Pokémon from its URL.
    Extracts ID, Name, Types, Abilities, Base Stats, and Image URLs.
    """
    try:
        response = requests.get(url, timeout=10) # Timeout for individual requests
        response.raise_for_status()
        data = response.json()

        # Extracting desired fields. We'll handle potential missing keys gracefully.
        pokemon_info = {
            "id": data.get("id"),
            "name": data.get("name"),
            "types": [t["type"]["name"] for t in data.get("types", [])],
            "abilities": [
                {
                    "name": a["ability"]["name"],
                    "is_hidden": a["is_hidden"]
                }
                for a in data.get("abilities", [])
            ],
            "stats": {
                s["stat"]["name"]: s["base_stat"]
                for s in data.get("stats", [])
            },
            "sprites": {
                "front_default": data.get("sprites", {}).get("front_default"),
                "official_artwork": data.get("sprites", {}).get("other", {}).get("official-artwork", {}).get("front_default")
            }
        }
        return pokemon_info
    except requests.exceptions.RequestException as e:
        print(f"Error fetching {url}: {e}")
        return None
    except KeyError as e:
        print(f"KeyError in JSON for {url}: {e}. Data structure might be unexpected.")
        return None
   except Exception as e:
        print(f"An unexpected error occurred for {url}: {e}")
        return None

def main():
    start_time = time.time()

    pokemon_urls = fetch_pokemon_list()
    if not pokemon_urls:
        print("No Pokémon URLs to process. Exiting.")
        return

    all_pokemon_details = []
    
    # Using ThreadPoolExecutor for concurrent fetching
    # This creates a pool of 'MAX_CONCURRENT_REQUESTS' threads to fetch data.
    with ThreadPoolExecutor(max_workers=MAX_CONCURRENT_REQUESTS) as executor:
        # Submit tasks to the executor
        # future_to_url maps each Future object back to the URL it was fetching, useful for debugging
        future_to_url = {executor.submit(fetch_pokemon_detail, url): url for url in pokemon_urls}
        
        # as_completed yields futures as they complete
        for i, future in enumerate(as_completed(future_to_url)):
            url = future_to_url[future]
            try:
                pokemon_data = future.result() # Get the result of the task
                if pokemon_data:
                    all_pokemon_details.append(pokemon_data)
                    # Optional: Print progress
                    if (i + 1) % 100 == 0 or (i + 1) == len(pokemon_urls):
                        print(f"Processed {i + 1}/{len(pokemon_urls)} Pokémon details...")
            except Exception as exc:
                print(f'{url} generated an exception: {exc}')
    
    end_time = time.time()
    print(f"\nFetched details for {len(all_pokemon_details)} Pokémon in {end_time - start_time:.2f} seconds.")

    # Optional: Dump all collected data to a single JSON file
    output_filename = "all_pokemon_data.json"
    try:
        with open(output_filename, 'w', encoding='utf-8') as f:
            json.dump(all_pokemon_details, f, indent=4, ensure_ascii=False)
        print(f"All collected Pokémon data dumped to '{output_filename}'")
    except Exception as e:
        print(f"Error dumping data to file: {e}")

if __name__ == "__main__":
    main()
