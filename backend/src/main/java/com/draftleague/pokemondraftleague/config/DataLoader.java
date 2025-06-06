// src/main/java/com/draftleague/pokemondraftleague/config/DataLoader.java
package com.draftleague.pokemondraftleague.config;

import com.draftleague.pokemondraftleague.model.Pokemon;
import com.draftleague.pokemondraftleague.repository.PokemonRepository; // <<< CHANGED THIS LINE (removed 's')
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;
import org.springframework.core.io.ClassPathResource;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

@Component
public class DataLoader implements CommandLineRunner {

	private final PokemonRepository pokemonRepository;
	private final Random random = new Random();

	public DataLoader(PokemonRepository pokemonRepository) {
		this.pokemonRepository = pokemonRepository;
	}

	@Override
	public void run(String... args) throws Exception {
		if (pokemonRepository.count() == 0) {
			System.out.println("Loading Pokemon data from pokemon.csv...");
			loadPokemonFromCsv();
			System.out.println("Pokemon data loaded successfully: " + pokemonRepository.count() + " Pokemon added.");
		} else {
			System.out.println("Pokemon data already exists in the database. Skipping CSV load.");
		}
	}

	private void loadPokemonFromCsv() {
		List<Pokemon> pokemonList = new ArrayList<>();
		try (BufferedReader reader = new BufferedReader(new InputStreamReader(
				new ClassPathResource("pokemon.csv").getInputStream()))) {

			String line;
			reader.readLine(); // Skip the header row

			while ((line = reader.readLine()) != null) {
				String[] data = line.split(",");

				if (data.length != 13) { // Ensure correct number of columns for your CSV
					System.err.println("Skipping malformed line (incorrect number of columns): " + line);
					continue;
				}

				try {
					Pokemon pokemon = new Pokemon();
					pokemon.setId(Integer.parseInt(data[0].trim()));
					pokemon.setName(data[1].trim());
					pokemon.setForm(data[2].trim().isEmpty() ? null : data[2].trim());
					pokemon.setType1(data[3].trim());
					pokemon.setType2(data[4].trim().isEmpty() ? null : data[4].trim());

					pokemon.setTotal(Integer.parseInt(data[5].trim()));
					pokemon.setHp(Integer.parseInt(data[6].trim()));
					pokemon.setAttack(Integer.parseInt(data[7].trim()));
					pokemon.setDefense(Integer.parseInt(data[8].trim()));
					pokemon.setSpAttack(Integer.parseInt(data[9].trim()));
					pokemon.setSpDefense(Integer.parseInt(data[10].trim()));
					pokemon.setSpeed(Integer.parseInt(data[11].trim()));
					pokemon.setGeneration(Integer.parseInt(data[12].trim()));

					int randomCost = random.nextInt(18) + 3;
					pokemon.setDraftCost(randomCost);

					pokemonList.add(pokemon);
				} catch (NumberFormatException e) {
					System.err.println("Error parsing number in line: " + line + " - " + e.getMessage());
				} catch (ArrayIndexOutOfBoundsException e) { // Added this catch for robustness
					System.err.println("Not enough columns in line: " + line + " - " + e.getMessage());
				}
			}
			pokemonRepository.saveAll(pokemonList);
		} catch (Exception e) {
			System.err.println("Failed to load Pokemon from CSV: " + e.getMessage());
			e.printStackTrace();
		}
	}
}
