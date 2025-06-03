// src/main/java/com/draftleague/pokemondraftleague/model/Pokemon.java
package com.draftleague.pokemondraftleague.model;

import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;

import jakarta.persistence.ManyToOne;
import jakarta.persistence.JoinColumn;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.ToString;

@Entity
@Data
@NoArgsConstructor // Lombok: Generates a no-argument constructor (required by JPA)
@AllArgsConstructor // Lombok: Generates a constructor with arguments for all fields
@ToString(exclude = { "draftedByTrainer" }) // Exclude Trainer from ToString to prevent infinite loops
@EqualsAndHashCode(exclude = { "draftedByTrainer" }) // Exclude Trainer from EqualsAndHashCode
public class Pokemon {

    @Id // Marks this field as the primary key of the table
    @GeneratedValue(strategy = GenerationType.IDENTITY) // Tells the DB to auto-increment this ID
    private Long id;
    private String name;
    private String type1; // e.g., "Grass", "Fire"
    private String type2; // e.g., "Poison", "Flying" (can be null)

    // Base Stat Spread
    private Integer baseHp;
    private Integer baseAttack;
    private Integer baseDefense;
    private Integer baseSpecialAttack;
    private Integer baseSpecialDefense;
    private Integer baseSpeed;

    private Integer draftCost; // Field for the cost of the Pokemon in the draft

    // --- Relationship ---

    // A Pokemon is drafted by ONE Trainer (or null if not drafted yet)
    // Many-to-One relationship from Pokemon to Trainer
    @ManyToOne
    @JoinColumn(name = "trainer_id") // Foreign key column in the 'pokemon' table. Nullable by default.
    private Trainer draftedByTrainer; // The trainer who drafted this Pokemon (can be null if not drafted)
}
