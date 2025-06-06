// src/main/java/com/draftleague/pokemondraftleague/repository/PokemonRepository.java
package com.draftleague.pokemondraftleague.repository;

import com.draftleague.pokemondraftleague.model.Pokemon;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository // Marks this interface as a Spring Data JPA repository
public interface PokemonRepository extends JpaRepository<Pokemon, Integer> {
	// Pokemon: The entity type this repository manages
	// Long: The data type of the entity's primary key (Pokemon's 'id' is a Long)

	// Example of a custom query method: find a Pokemon by its name
	Optional<Pokemon> findByName(String name);
}
