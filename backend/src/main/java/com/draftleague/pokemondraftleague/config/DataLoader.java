package com.draftleague.pokemondraftleague.config;

import com.draftleague.pokemondraftleague.model.Pokemon;
import com.draftleague.pokemondraftleague.model.Trainer; // Import Trainer model
import com.draftleague.pokemondraftleague.repository.PokemonRepository;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;
import jakarta.transaction.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

@Component
public class DataLoader implements CommandLineRunner {

	private final PokemonRepository pokemonRepository;

	@PersistenceContext
	private EntityManager entityManager;

	private final Random random = new Random();

	private static final int MIN_DRAFT_COST = 3;
	private static final int MAX_DRAFT_COST = 20;

	public DataLoader(PokemonRepository pokemonRepository) {
		this.pokemonRepository = pokemonRepository;
	}

	private Integer generateRandomDraftCost() {
		return random.nextInt(MAX_DRAFT_COST - MIN_DRAFT_COST + 1) + MIN_DRAFT_COST;
	}

	@Override
	@Transactional
	public void run(String... args) throws Exception {
		System.out.println("Checking for existing Pokemon data (in PostgreSQL)...");
		if (pokemonRepository.count() > 0) {
			System.out.println("Existing Pokemon data found. Deleting all existing data for a fresh hardcoded load.");
			pokemonRepository.deleteAllInBatch();
			entityManager.flush();
			entityManager.clear();
		} else {
			System.out.println("No existing Pokemon data found, proceeding with hardcoded load.");
		}

		System.out.println("Loading hardcoded Pokemon data directly for prototype with randomized costs and images...");

		List<Pokemon> initialPokemonData = new ArrayList<>();

		// Constructor parameters for Pokemon (based on your updated model):
		// (Integer id, String name, String type1, String type2, Integer hp, Integer
		// attack, Integer defense,
		// Integer spAttack, Integer spDefense, Integer speed, Integer total, String
		// ability, String hiddenAbility,
		// String form, Integer draftCost, Integer generation, Trainer draftedByTrainer)

		initialPokemonData.add(new Pokemon(null, "Bulbasaur", "Grass", "Poison", 45, 49, 49, 65, 65, 45, 318,
				"Overgrow", null, null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Charmander", "Fire", null, 39, 52, 43, 60, 50, 65, 309,
				"Blaze", null, null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Squirtle", "Water", null, 44, 48, 65, 50, 64, 43, 314,
				"Torrent", null, null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Pikachu", "Electric", null, 35, 55, 40, 50, 50, 90, 320,
				"Static", "Lightning Rod", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Mewtwo", "Psychic", null, 106, 110, 90, 154, 90, 130, 680,
				"Pressure", "Unnerve", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Eevee", "Normal", null, 55, 55, 50, 45, 65, 55, 325,
				"Run Away", "Adaptability", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Charizard", "Fire", "Flying", 78, 84, 78, 109, 85, 100, 534,
				"Blaze", "Solar Power", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Snorlax", "Normal", null, 160, 110, 65, 65, 110, 30, 540,
				"Immunity", "Thick Fat", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Gengar", "Ghost", "Poison", 60, 65, 60, 130, 75, 110, 500,
				"Cursed Body", null, null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Dragonite", "Dragon", "Flying", 91, 134, 95, 100, 100, 80, 600,
				"Inner Focus", "Multiscale", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Venusaur", "Grass", "Poison", 80, 82, 83, 100, 100, 80, 525,
				"Overgrow", "Chlorophyll", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Blastoise", "Water", null, 79, 83, 100, 85, 105, 78, 530,
				"Torrent", "Rain Dish", null, generateRandomDraftCost(), 1, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Umbreon", "Dark", null, 95, 65, 110, 60, 130, 65, 525,
				"Synchronize", "Inner Focus", null, generateRandomDraftCost(), 2, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Scizor", "Bug", "Steel", 70, 130, 100, 55, 80, 65, 500,
				"Swarm", "Light Metal", null, generateRandomDraftCost(), 2, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Tyranitar", "Rock", "Dark", 100, 134, 110, 95, 100, 61, 600,
				"Sand Stream", "Unnerve", null, generateRandomDraftCost(), 2, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Blaziken", "Fire", "Fighting", 80, 120, 70, 110, 70, 80, 530,
				"Blaze", "Speed Boost", null, generateRandomDraftCost(), 3, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Gardevoir", "Psychic", "Fairy", 68, 65, 65, 125, 115, 80, 518,
				"Synchronize", "Telepathy", null, generateRandomDraftCost(), 3, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Rayquaza", "Dragon", "Flying", 105, 150, 90, 150, 90, 95, 780,
				"Air Lock", null, null, generateRandomDraftCost(), 3, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Lucario", "Fighting", "Steel", 70, 110, 70, 115, 70, 90, 525,
				"Steadfast", "Inner Focus", null, generateRandomDraftCost(), 4, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Gallade", "Psychic", "Fighting", 68, 125, 65, 65, 115, 80, 518,
				"Steadfast", "Justified", null, generateRandomDraftCost(), 4, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Hydreigon", "Dark", "Dragon", 92, 105, 90, 125, 90, 98, 600,
				"Levitate", null, null, generateRandomDraftCost(), 5, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Zekrom", "Dragon", "Electric", 100, 150, 120, 120, 100, 90, 680,
				"Teravolt", null, null, generateRandomDraftCost(), 5, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Greninja", "Water", "Dark", 72, 95, 67, 103, 71, 122, 530,
				"Torrent", "Protean", null, generateRandomDraftCost(), 6, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Talonflame", "Fire", "Flying", 78, 81, 71, 74, 69, 126, 499,
				"Flame Body", "Gale Wings", null, generateRandomDraftCost(), 6, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Mimikyu", "Ghost", "Fairy", 55, 90, 80, 50, 105, 96, 470,
				"Disguise", null, null, generateRandomDraftCost(), 7, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Corviknight", "Flying", "Steel", 98, 87, 105, 53, 85, 67, 495,
				"Pressure", "Unnerve", null, generateRandomDraftCost(), 8, (Trainer) null));
		initialPokemonData.add(new Pokemon(null, "Garchomp", "Dragon", "Ground", 108, 130, 95, 80, 85, 102, 600,
				"Sand Veil", "Rough Skin", null, generateRandomDraftCost(), 4, (Trainer) null));

		for (Pokemon pokemon : initialPokemonData) {
			entityManager.persist(pokemon);
		}
		System.out.println(
				"Hardcoded Pokemon data saved to PostgreSQL: " + initialPokemonData.size() + " Pokemon added.");
	}
}
