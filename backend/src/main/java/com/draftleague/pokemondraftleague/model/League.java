// src/main/java/com/draftleague/pokemondraftleague/model/League.java
package com.draftleague.pokemondraftleague.model;

import jakarta.persistence.CascadeType; // For cascading operations
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.OneToMany; // For the one-to-many relationship
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode; // For Lombok
import lombok.NoArgsConstructor;
import lombok.ToString; // For Lombok

import java.util.Set; // Use Set for collections here

@Entity
@Data
@NoArgsConstructor
@AllArgsConstructor
@ToString(exclude = { "trainers", "matches" }) // Exclude collections to prevent infinite loops
@EqualsAndHashCode(exclude = { "trainers", "matches" }) // Exclude collections from EqualsAndHashCode
public class League {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    private String name;
    private String status; // e.g., "SETUP", "DRAFTING", "ACTIVE", "COMPLETED"
    private Integer maxPokemonPerTrainer; // e.g., 6, 10

    // --- Relationships ---

    // A League has MANY Trainers
    // 'mappedBy' indicates the field in the 'Trainer' entity that owns the
    // relationship
    @OneToMany(mappedBy = "league", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Trainer> trainers;

    // A League has MANY Matches
    // 'mappedBy' indicates the field in the 'Match' entity that owns the
    // relationship
    @OneToMany(mappedBy = "league", cascade = CascadeType.ALL, orphanRemoval = true)
    private Set<Match> matches;

    // Custom constructor for initial seeding (without collections)
    public League(Long id, String name, String status, Integer maxPokemonPerTrainer) {
        this.id = id;
        this.name = name;
        this.status = status;
        this.maxPokemonPerTrainer = maxPokemonPerTrainer;
    }
}
