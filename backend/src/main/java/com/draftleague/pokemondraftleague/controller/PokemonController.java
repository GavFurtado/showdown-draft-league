// src/main/java/com/draftleague/pokemondraftleague/controller/PokemonController.java
package com.draftleague.pokemondraftleague.controller;

import com.draftleague.pokemondraftleague.model.Pokemon;
import com.draftleague.pokemondraftleague.service.PokemonService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/pokemon")
public class PokemonController {

	private final PokemonService pokemonService;

	public PokemonController(PokemonService pokemonService) {
		this.pokemonService = pokemonService;
	}

	@GetMapping
	public ResponseEntity<List<Pokemon>> getAllPokemon() {
		List<Pokemon> pokemonList = pokemonService.getAllPokemon();
		return ResponseEntity.ok(pokemonList);
	}

	@GetMapping("/{id}")
	public ResponseEntity<Pokemon> getPokemonById(@PathVariable Long id) {
		return pokemonService.getPokemonById(id)
				.map(ResponseEntity::ok)
				.orElseGet(() -> new ResponseEntity<>(HttpStatus.NOT_FOUND));
	}

	// You might not need these for a draft league if Pokemon data is seeded
	// @PostMapping
	// public ResponseEntity<Pokemon> createPokemon(@RequestBody Pokemon pokemon) {
	// Pokemon savedPokemon = pokemonService.savePokemon(pokemon);
	// return new ResponseEntity<>(savedPokemon, HttpStatus.CREATED);
	// }

	// @DeleteMapping("/{id}")
	// public ResponseEntity<Void> deletePokemon(@PathVariable Long id) {
	// pokemonService.deletePokemon(id);
	// return new ResponseEntity<>(HttpStatus.NO_CONTENT);
	// }
}
