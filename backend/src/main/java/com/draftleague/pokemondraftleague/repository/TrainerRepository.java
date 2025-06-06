// src/main/java/com/draftleague/pokemondraftleague/repository/TrainerRepository.java
package com.draftleague.pokemondraftleague.repository;

import com.draftleague.pokemondraftleague.model.Trainer;
import com.draftleague.pokemondraftleague.model.League;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List; // For findByLeague
import java.util.Optional;

@Repository
public interface TrainerRepository extends JpaRepository<Trainer, Long> {
    // find Trainer by their username
    Optional<Trainer> findByUsername(String username);

    // find all Trainers of a League
    List<Trainer> findByLeague(League league);
}
