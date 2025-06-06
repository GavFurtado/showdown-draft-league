// src/main/java/com/draftleague/pokemondraftleague/service/PokemonService.java
package com.draftleague.pokemondraftleague.service;

import com.draftleague.pokemondraftleague.model.Pokemon;
import com.draftleague.pokemondraftleague.repository.PokemonRepository;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class PokemonService {

	private final PokemonRepository pokemonRepository;

	public PokemonService(PokemonRepository pokemonRepository) {
		this.pokemonRepository = pokemonRepository;
	}

	public Pokemon savePokemon(Pokemon pokemon) {
		return pokemonRepository.save(pokemon);
	}

	public List<Pokemon> getAllPokemon() {
		return pokemonRepository.findAll();
	}

	public Optional<Pokemon> getPokemonById(Integer id) { // Changed Long to Integer
		return pokemonRepository.findById(id);
	}

	public void deletePokemon(Integer id) { // Changed Long to Integer
		pokemonRepository.deleteById(id);
	}

	public Optional<Pokemon> getPokemonByName(String name) {
		return pokemonRepository.findByName(name);
	}
}
