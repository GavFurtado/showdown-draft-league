// src/main/java/com/draftleague/pokemondraftleague/controller/PokemonController.java
package com.draftleague.pokemondraftleague.controller;

import com.draftleague.pokemondraftleague.model.Pokemon;
import com.draftleague.pokemondraftleague.service.PokemonService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/pokemon")
public class PokemonController {

	private final PokemonService pokemonService;

	public PokemonController(PokemonService pokemonService) {
		this.pokemonService = pokemonService;
	}

	@PostMapping
	public ResponseEntity<Pokemon> createPokemon(@RequestBody Pokemon pokemon) {
		Pokemon savedPokemon = pokemonService.savePokemon(pokemon);
		return new ResponseEntity<>(savedPokemon, HttpStatus.CREATED);
	}

	@GetMapping
	public ResponseEntity<List<Pokemon>> getAllPokemon() {
		List<Pokemon> pokemonList = pokemonService.getAllPokemon();
		return new ResponseEntity<>(pokemonList, HttpStatus.OK);
	}

	@GetMapping("/{id}")
	public ResponseEntity<Pokemon> getPokemonById(@PathVariable Integer id) { // Changed Long to Integer
		return pokemonService.getPokemonById(id)
				.map(pokemon -> new ResponseEntity<>(pokemon, HttpStatus.OK))
				.orElse(new ResponseEntity<>(HttpStatus.NOT_FOUND));
	}

	@GetMapping("/name/{name}")
	public ResponseEntity<Pokemon> getPokemonByName(@PathVariable String name) {
		return pokemonService.getPokemonByName(name)
				.map(pokemon -> new ResponseEntity<>(pokemon, HttpStatus.OK))
				.orElse(new ResponseEntity<>(HttpStatus.NOT_FOUND));
	}

	@PutMapping("/{id}")
	public ResponseEntity<Pokemon> updatePokemon(@PathVariable Integer id, @RequestBody Pokemon pokemon) {
		if (!pokemon.getId().equals(id)) {
			return new ResponseEntity<>(HttpStatus.BAD_REQUEST);
		}
		if (pokemonService.getPokemonById(id).isEmpty()) {
			return new ResponseEntity<>(HttpStatus.NOT_FOUND);
		}
		Pokemon updatedPokemon = pokemonService.savePokemon(pokemon);
		return new ResponseEntity<>(updatedPokemon, HttpStatus.OK);
	}

	@DeleteMapping("/{id}")
	public ResponseEntity<Void> deletePokemon(@PathVariable Integer id) { // Changed Long to Integer
		if (pokemonService.getPokemonById(id).isEmpty()) {
			return new ResponseEntity<>(HttpStatus.NOT_FOUND);
		}
		pokemonService.deletePokemon(id);
		return new ResponseEntity<>(HttpStatus.NO_CONTENT);
	}
}
