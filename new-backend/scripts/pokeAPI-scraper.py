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

        # Extracting desired fields
        pokemon_id = data.get("id")
        pokemon_name = data.get("name")

        # Get the species URL from the initial pokemon data
        species_url = data.get("species", {}).get("url")
        species_data = None
        try:
            species_response = requests.get(species_url, timeout=10)
            species_response.raise_for_status()
            species_data = species_response.json()
        except requests.exceptions.RequestException as e:
            print(f"Error fetching species data for {pokemon_name} ({pokemon_id}): {e}")

        dex_id = species_data.get("id") if species_data else None

        pokemon_info = {
            "id": pokemon_id,
            "dex_id": dex_id,
            "name": pokemon_name,
            "types": [t["type"]["name"] for t in data.get("types", [])],
            "abilities": [
                {
                    "name": a["ability"]["name"],
                    "is_hidden": a["is_hidden"]
                }
                for a in data.get("abilities", [])
            ],
            "stats": {
                "hp": next((s["base_stat"] for s in data.get("stats", []) if s["stat"]["name"] == "hp"), 0),
                "attack": next((s["base_stat"] for s in data.get("stats", []) if s["stat"]["name"] == "attack"), 0),
                "defense": next((s["base_stat"] for s in data.get("stats", []) if s["stat"]["name"] == "defense"), 0),
                "special_attack": next((s["base_stat"] for s in data.get("stats", []) if s["stat"]["name"] == "special-attack"), 0),
                "special_defense": next((s["base_stat"] for s in data.get("stats", []) if s["stat"]["name"] == "special-defense"), 0),
                "speed": next((s["base_stat"] for s in data.get("stats", []) if s["stat"]["name"] == "speed"), 0)
            },
            "sprites": {
                "front_default": data.get("sprites", {}).get("front_default"),
                "official_artwork": data.get("sprites", {}).get("other", {}).get("official-artwork", {}).get("front_default")
            }
        }

        # Prioritize official artwork if available, otherwise use front_default
        primary_sprite = pokemon_info["sprites"]["official_artwork"]
        if not primary_sprite:
            primary_sprite = pokemon_info["sprites"]["front_default"]

        # Fallback to a generic placeholder if no sprite is found
        if not primary_sprite:
            primary_sprite = "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/0.png" # Generic placeholder (e.g., a question mark or empty sprite)

        pokemon_info["sprites"]["front_default"] = primary_sprite
        pokemon_info["sprites"]["official_artwork"] = primary_sprite

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

    # Sort the collected data by 'id' before dumping
    print(f"Sorting {len(all_pokemon_details)} Pokémon by ID before dumping...")
    sorted_pokemon_data = sorted(all_pokemon_details, key=lambda p: p.get("id", float('inf')))

    # Dump all collected and sorted data to a single JSON file
    output_filename = "data/all_pokemon_data.json"
    try:
        # Ensure the 'data' directory exists
        os.makedirs(os.path.dirname(output_filename), exist_ok=True)
        with open(output_filename, 'w', encoding='utf-8') as f:
            json.dump(sorted_pokemon_data, f, indent=4, ensure_ascii=False)
        print(f"All collected and sorted Pokémon data dumped to '{output_filename}'")
    except Exception as e:
        print(f"Error dumping data to file: {e}")

if __name__ == "__main__":
    main()
