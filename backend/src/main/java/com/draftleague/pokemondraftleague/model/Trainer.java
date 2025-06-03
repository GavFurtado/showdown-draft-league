// src/main/java/com/draftleague/pokemondraftleague/model/Trainer.java
package com.draftleague.pokemondraftleague.model;

import jakarta.persistence.CascadeType; // For cascading operations
import jakarta.persistence.CollectionTable; // For ElementCollection (if used, not for Trainer itself)
import jakarta.persistence.Column;
import jakarta.persistence.ElementCollection; // For collections of basic types
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn; // For foreign key
import jakarta.persistence.ManyToOne; // For many-to-one relationship
import jakarta.persistence.OneToMany; // For one-to-many relationship
import lombok.AllArgsConstructor;
import lombok.Data; // Provides getters, setters, equals, hashCode, toString
import lombok.EqualsAndHashCode; // Custom equals/hashCode generation
import lombok.NoArgsConstructor; // Provides no-argument constructor
import lombok.ToString; // Custom toString generation

import java.util.Set; // For the collection of drafted Pokemon

@Entity // This annotation marks this class as a JPA entity, meaning it maps to a
        // database table.
@Data
@NoArgsConstructor
@AllArgsConstructor
@ToString(exclude = { "draftedPokemon" }) // IMPORTANT: Exclude collections to prevent StackOverflowError in toString()
                                          // for bi-directional relationships.
@EqualsAndHashCode(exclude = { "draftedPokemon" })
public class Trainer {

    @Id // Marks this field as the primary key of the table.
    @GeneratedValue(strategy = GenerationType.IDENTITY) // Tells the database to auto-increment this ID.
    private Long id;

    private String name; // Trainer's display name.
    private String discordId; // Optional: Discord user ID if you integrate with Discord.

    @Column(unique = true, nullable = false) // Ensures username is unique and cannot be null.
    private String username; // Username for login.

    @Column(nullable = false) // Ensures password cannot be null. This will store the BCrypt hashed password.
    private String password; // Stores the hashed password.

    // --- Relationships ---

    // Many Trainers can belong to ONE League.
    // This is the owning side of the ManyToOne relationship, meaning the 'trainer'
    // table will have the foreign key.
    @ManyToOne
    @JoinColumn(name = "league_id", nullable = false) // Creates a 'league_id' column in the 'trainer' table that
                                                      // references the League's ID. Must not be null.
    private League league; // The League this Trainer belongs to.

    // One Trainer can draft MANY Pokemon.
    // This is the inverse (non-owning) side of the relationship, as the 'Pokemon'
    // entity owns the 'trainer_id' foreign key.
    @OneToMany(mappedBy = "draftedByTrainer", // 'mappedBy' refers to the field name in the Pokemon entity that holds
                                              // the Trainer reference.
            cascade = CascadeType.ALL, // If a Trainer is deleted, cascade this operation to their drafted Pokemon.
            orphanRemoval = true) // If a Pokemon is removed from this set, it's also removed from the database.
    private Set<Pokemon> draftedPokemon; // A collection of Pokemon this trainer has drafted.
}
