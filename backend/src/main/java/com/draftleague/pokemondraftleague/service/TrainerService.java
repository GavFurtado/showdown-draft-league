// src/main/java/com/draftleague/pokemondraftleague/service/TrainerService.java
package com.draftleague.pokemondraftleague.service;

import com.draftleague.pokemondraftleague.model.Trainer;
import com.draftleague.pokemondraftleague.model.League;
import com.draftleague.pokemondraftleague.repository.TrainerRepository;
import com.draftleague.pokemondraftleague.repository.LeagueRepository;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional; // For transactional methods

import java.util.List;
import java.util.Optional;

@Service
public class TrainerService {

	private final TrainerRepository trainerRepository;
	private final LeagueRepository leagueRepository; // To assign trainer to a league
	private final BCryptPasswordEncoder passwordEncoder; // Injected for hashing

	public TrainerService(TrainerRepository trainerRepository, LeagueRepository leagueRepository,
			BCryptPasswordEncoder passwordEncoder) {
		this.trainerRepository = trainerRepository;
		this.leagueRepository = leagueRepository;
		this.passwordEncoder = passwordEncoder;
	}

	public List<Trainer> getAllTrainers() {
		return trainerRepository.findAll();
	}

	public Optional<Trainer> getTrainerById(Long id) {
		return trainerRepository.findById(id);
	}

	/**
	 * Registers a new trainer, hashing their password.
	 * 
	 * @param trainer  The trainer object with plain-text password.
	 * @param leagueId The ID of the league the trainer should join.
	 * @return The saved Trainer object with hashed password.
	 * @throws IllegalArgumentException if username exists or league not found.
	 */
	@Transactional
	public Trainer registerTrainer(Trainer trainer, Long leagueId) {
		if (trainerRepository.findByUsername(trainer.getUsername()).isPresent()) {
			throw new IllegalArgumentException("Username already exists.");
		}
		Optional<League> leagueOptional = leagueRepository.findById(leagueId);
		if (leagueOptional.isEmpty()) {
			throw new IllegalArgumentException("League with ID " + leagueId + " not found.");
		}

		trainer.setPassword(passwordEncoder.encode(trainer.getPassword())); // Hash the password
		trainer.setLeague(leagueOptional.get()); // Assign the league

		return trainerRepository.save(trainer);
	}

	/**
	 * Authenticates a trainer by checking username and password.
	 * 
	 * @param username The trainer's username.
	 * @param password The plain-text password provided by the user.
	 * @return An Optional containing the Trainer if authentication is successful,
	 *         otherwise empty.
	 */
	public Optional<Trainer> loginTrainer(String username, String password) {
		Optional<Trainer> trainerOptional = trainerRepository.findByUsername(username);
		if (trainerOptional.isPresent()) {
			Trainer trainer = trainerOptional.get();
			// Compare the provided plain-text password with the stored hashed password
			if (passwordEncoder.matches(password, trainer.getPassword())) {
				return Optional.of(trainer);
			}
		}
		return Optional.empty(); // Authentication failed
	}

	public void deleteTrainer(Long id) {
		trainerRepository.deleteById(id);
	}
}
