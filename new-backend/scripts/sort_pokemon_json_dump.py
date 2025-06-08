import json
import os

def sort_pokemon_data_in_place(filepath="all_pokemon_data.json"):
    """
    Loads Pokémon data from a JSON file, sorts it by 'id', and
    overwrites the original file with the sorted data.

    Args:
        filepath (str): The path to the JSON file containing Pokémon data.
    """
    if not os.path.exists(filepath):
        print(f"Error: File not found at '{filepath}'")
        return

    print(f"Loading data from '{filepath}'...")
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            pokemon_data = json.load(f) # Read the entire file into memory
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON from '{filepath}': {e}")
        return
    except Exception as e:
        print(f"An unexpected error occurred while reading the file: {e}")
        return

    if not isinstance(pokemon_data, list):
        print(f"Error: Expected a JSON array (list) at the root of '{filepath}', but found {type(pokemon_data).__name__}.")
        print("Please ensure the file contains a list of Pokémon objects.")
        return

    print(f"Sorting {len(pokemon_data)} Pokémon by ID...")

    # Sort the list of dictionaries by the 'id' key.
    # We use .get("id", float('inf')) to safely handle any entries that might
    # unexpectedly lack an 'id' by placing them at the end.
    sorted_pokemon_data = sorted(pokemon_data, key=lambda p: p.get("id", float('inf')))

    print(f"Overwriting '{filepath}' with sorted data...")
    try:
        with open(filepath, 'w', encoding='utf-8') as f: # Use 'w' mode to overwrite
            json.dump(sorted_pokemon_data, f, indent=4, ensure_ascii=False)
        print("Sorting complete. File overwritten successfully.")
    except Exception as e:
        print(f"Error writing sorted data back to '{filepath}': {e}")

if __name__ == "__main__":
    # Assuming 'all_pokemon_data.json' is in the same directory as this script.
    # If it's in a 'data' subdirectory, specify the path like:
    # file_to_sort = os.path.join("data", "all_pokemon_data.json")

    
    file_to_sort = "../data/all_pokemon_data.json" # This will be both input and output
    if not os.path.exists(file_to_sort):
        file_to_sort = "./data/all_pokemon_data.json"

    sort_pokemon_data_in_place(file_to_sort)

    # Optional: Verify the first few entries in the now-sorted file
    if os.path.exists(file_to_sort):
        print("\nVerifying first 5 entries in the now-sorted file:")
        try:
            with open(file_to_sort, 'r', encoding='utf-8') as f:
                sorted_data = json.load(f)
                for i, pokemon in enumerate(sorted_data[:5]):
                    print(f"  {i+1}: ID={pokemon.get('id')}, Name={pokemon.get('name')}")
        except Exception as e:
            print(f"Error during verification: {e}")
