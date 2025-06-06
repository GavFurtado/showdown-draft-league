package com.draftleague.pokemondraftleague.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

@Configuration
@EnableWebSecurity
public class SecurityConfig {

	@Bean
	public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
		http
				.csrf(AbstractHttpConfigurer::disable) // Disable CSRF for API
				.authorizeHttpRequests(authorize -> authorize
						// [IMP]: Permit ALL requests ---
						.anyRequest().permitAll() // Allow all requests without authentication
				);
		// Remove or comment out any formLogin or httpBasic lines if present,
		// as they would imply authentication flows.
		// http.formLogin(withDefaults()); // Not needed if everything is permitted
		// http.httpBasic(withDefaults()); // Not needed if everything is permitted

		return http.build();
	}

	@Bean
	public BCryptPasswordEncoder bCryptPasswordEncoder() {
		return new BCryptPasswordEncoder();
	}
}
