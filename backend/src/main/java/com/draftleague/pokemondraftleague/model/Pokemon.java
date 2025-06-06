package com.draftleague.pokemondraftleague.model;

import com.fasterxml.jackson.annotation.JsonBackReference;
import com.fasterxml.jackson.annotation.JsonManagedReference;

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
@NoArgsConstructor
@AllArgsConstructor // Lombok: Generates a constructor with arguments for all fields
@ToString(exclude = { "draftedByTrainer" })
@EqualsAndHashCode(exclude = { "draftedByTrainer" })
public class Pokemon {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;
    private String name;
    private String type1;
    private String type2; // (can be null)

    // Base Stat Spread (order adjusted to align with common CSV order/parsing)
    private Integer hp;
    private Integer attack;
    private Integer defense;
    private Integer spAttack; // Renamed from specialAttack
    private Integer spDefense; // Renamed from specialDefense
    private Integer speed;
    private Integer total;

    // Abilities (added these fields as they are parsed from CSV)
    private String ability;
    private String hiddenAbility; // (can be null)

    // Other fields not directly from CSV or for specific game mechanics
    private String form; // (can be null)
    private Integer draftCost; // Field for the cost of the Pokemon in the draft (can be null initially)
    private Integer generation; // (can be null, or default to 0 if preferred)

    // --- Relationship ---
    @ManyToOne
    @JoinColumn(name = "trainer_id")
    @JsonBackReference
    private Trainer draftedByTrainer; // The trainer who drafted this Pokemon (can be null if not drafted)
}
