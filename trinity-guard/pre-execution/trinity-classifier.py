#!/usr/bin/env python3
"""
Trinity Classifier - Springfield's Strategic Analysis Engine
カフェ・ズッケロ Trinity Guard System v2.0

This module analyzes user inputs and determines whether the Trinity system
should be activated based on complexity, scope, and strategic importance.

Author: Springfield (The Strategic Architect)
Role: User input analysis and Trinity activation decision
"""

import re
import json
import sys
import os
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
from enum import Enum

class ComplexityLevel(Enum):
    SIMPLE = 1
    MODERATE = 2
    COMPLEX = 3
    TRINITY_REQUIRED = 4

class TrinityDomain(Enum):
    STRATEGY = "springfield-strategist"
    TECHNICAL = "krukai-optimizer"
    SECURITY = "vector-auditor"
    INTEGRATED = "trinitas-coordinator"

@dataclass
class AnalysisResult:
    complexity: ComplexityLevel
    domains: List[TrinityDomain]
    confidence: float
    reasoning: str
    trinity_required: bool
    recommended_agent: Optional[str] = None

class TrinityClassifier:
    """
    Springfield's strategic analysis engine for determining Trinity activation.
    
    This classifier examines user input through multiple dimensions:
    - Technical complexity indicators
    - Multi-domain requirements
    - Risk assessment needs
    - Strategic importance signals
    """
    
    def __init__(self):
        self.complexity_patterns = {
            # Architecture and Design Patterns
            ComplexityLevel.TRINITY_REQUIRED: [
                r'(?i)(architecture|design\s+pattern|system\s+design)',
                r'(?i)(microservice|distributed|scalab)',
                r'(?i)(performance.*optimization|optimize.*performance)',
                r'(?i)(security.*review|security.*audit)',
                r'(?i)(multi.*tier|multi.*layer)',
                r'(?i)(integration.*strategy|api.*design)',
            ],
            
            # Complex Technical Tasks
            ComplexityLevel.COMPLEX: [
                r'(?i)(algorithm|data\s+structure)',
                r'(?i)(database.*design|schema.*design)',
                r'(?i)(concurrent|parallel|async)',
                r'(?i)(memory.*management|garbage.*collect)',
                r'(?i)(cache|caching|redis)',
                r'(?i)(deployment|devops|ci/cd)',
            ],
            
            # Moderate Complexity
            ComplexityLevel.MODERATE: [
                r'(?i)(refactor|clean.*code)',
                r'(?i)(test|testing|unit.*test)',
                r'(?i)(debug|debugging|troubleshoot)',
                r'(?i)(config|configuration)',
                r'(?i)(document|documentation)',
            ],
            
            # Simple Tasks
            ComplexityLevel.SIMPLE: [
                r'(?i)(fix.*typo|syntax.*error)',
                r'(?i)(simple.*function|basic.*implementation)',
                r'(?i)(hello.*world|tutorial)',
            ]
        }
        
        self.domain_patterns = {
            TrinityDomain.STRATEGY: [
                r'(?i)(project.*plan|roadmap|strategy)',
                r'(?i)(team.*coordination|resource.*allocation)',
                r'(?i)(architecture.*decision|technical.*debt)',
                r'(?i)(maintainability|scalability.*plan)',
            ],
            
            TrinityDomain.TECHNICAL: [
                r'(?i)(performance|optimization|efficiency)',
                r'(?i)(algorithm|implementation|code.*quality)',
                r'(?i)(memory|cpu|benchmark)',
                r'(?i)(compiler|build.*system)',
            ],
            
            TrinityDomain.SECURITY: [
                r'(?i)(security|vulnerability|penetration)',
                r'(?i)(authentication|authorization|encrypt)',
                r'(?i)(attack.*vector|threat.*model)',
                r'(?i)(compliance|audit|risk)',
            ],
        }
        
        # Keywords that always trigger Trinity mode
        self.trinity_keywords = [
            'trinity', 'trinitas', 'three.*perspective', 'multi.*domain',
            'comprehensive.*analysis', 'strategic.*technical.*security',
            'springfield.*krukai.*vector', 'tri-core'
        ]
        
    def analyze_input(self, user_input: str, context: Dict = None) -> AnalysisResult:
        """
        Analyze user input and determine Trinity activation requirements.
        
        Args:
            user_input: The user's request or query
            context: Additional context information
            
        Returns:
            AnalysisResult with complexity assessment and recommendations
        """
        if not user_input or len(user_input.strip()) < 5:
            return AnalysisResult(
                complexity=ComplexityLevel.SIMPLE,
                domains=[],
                confidence=1.0,
                reasoning="Input too short for complex analysis",
                trinity_required=False
            )
        
        # Check for explicit Trinity activation keywords
        trinity_explicit = self._check_trinity_keywords(user_input)
        if trinity_explicit:
            return AnalysisResult(
                complexity=ComplexityLevel.TRINITY_REQUIRED,
                domains=[TrinityDomain.INTEGRATED],
                confidence=1.0,
                reasoning="Explicit Trinity activation requested",
                trinity_required=True,
                recommended_agent="trinitas-coordinator"
            )
        
        # Analyze complexity level
        complexity = self._assess_complexity(user_input)
        
        # Identify involved domains
        domains = self._identify_domains(user_input)
        
        # Calculate confidence score
        confidence = self._calculate_confidence(user_input, complexity, domains)
        
        # Determine if Trinity is required
        trinity_required = self._should_activate_trinity(complexity, domains, confidence)
        
        # Generate reasoning
        reasoning = self._generate_reasoning(complexity, domains, confidence)
        
        # Recommend specific agent
        recommended_agent = self._recommend_agent(complexity, domains, trinity_required)
        
        return AnalysisResult(
            complexity=complexity,
            domains=domains,
            confidence=confidence,
            reasoning=reasoning,
            trinity_required=trinity_required,
            recommended_agent=recommended_agent
        )
    
    def _check_trinity_keywords(self, text: str) -> bool:
        """Check if text contains explicit Trinity activation keywords."""
        for keyword_pattern in self.trinity_keywords:
            if re.search(keyword_pattern, text, re.IGNORECASE):
                return True
        return False
    
    def _assess_complexity(self, text: str) -> ComplexityLevel:
        """Assess the complexity level of the request."""
        complexity_scores = {}
        
        for level, patterns in self.complexity_patterns.items():
            score = 0
            for pattern in patterns:
                matches = len(re.findall(pattern, text, re.IGNORECASE))
                score += matches
            complexity_scores[level] = score
        
        # Determine highest scoring complexity level
        if complexity_scores[ComplexityLevel.TRINITY_REQUIRED] > 0:
            return ComplexityLevel.TRINITY_REQUIRED
        elif complexity_scores[ComplexityLevel.COMPLEX] > 1:
            return ComplexityLevel.COMPLEX
        elif complexity_scores[ComplexityLevel.MODERATE] > 0:
            return ComplexityLevel.MODERATE
        else:
            return ComplexityLevel.SIMPLE
    
    def _identify_domains(self, text: str) -> List[TrinityDomain]:
        """Identify which Trinity domains are relevant to the request."""
        identified_domains = []
        
        for domain, patterns in self.domain_patterns.items():
            for pattern in patterns:
                if re.search(pattern, text, re.IGNORECASE):
                    identified_domains.append(domain)
                    break
        
        # If multiple domains detected, suggest integrated approach
        if len(identified_domains) >= 2:
            identified_domains.append(TrinityDomain.INTEGRATED)
        
        return list(set(identified_domains))  # Remove duplicates
    
    def _calculate_confidence(self, text: str, complexity: ComplexityLevel, 
                            domains: List[TrinityDomain]) -> float:
        """Calculate confidence score for the analysis."""
        base_confidence = 0.7
        
        # Boost confidence for clear indicators
        word_count = len(text.split())
        if word_count > 50:
            base_confidence += 0.1
        
        if complexity == ComplexityLevel.TRINITY_REQUIRED:
            base_confidence += 0.2
        
        if len(domains) > 1:
            base_confidence += 0.1
        
        return min(1.0, base_confidence)
    
    def _should_activate_trinity(self, complexity: ComplexityLevel, 
                               domains: List[TrinityDomain], 
                               confidence: float) -> bool:
        """Determine if Trinity system should be activated."""
        if complexity == ComplexityLevel.TRINITY_REQUIRED:
            return True
        
        if len(domains) >= 2 and confidence > 0.8:
            return True
        
        if complexity == ComplexityLevel.COMPLEX and len(domains) >= 1:
            return True
        
        return False
    
    def _generate_reasoning(self, complexity: ComplexityLevel, 
                          domains: List[TrinityDomain], 
                          confidence: float) -> str:
        """Generate human-readable reasoning for the decision."""
        reasoning_parts = []
        
        reasoning_parts.append(f"Complexity assessed as {complexity.name.lower()}")
        
        if domains:
            domain_names = [d.name.lower() for d in domains if d != TrinityDomain.INTEGRATED]
            if domain_names:
                reasoning_parts.append(f"Involves {', '.join(domain_names)} domains")
        
        reasoning_parts.append(f"Analysis confidence: {confidence:.1%}")
        
        return ". ".join(reasoning_parts)
    
    def _recommend_agent(self, complexity: ComplexityLevel, 
                        domains: List[TrinityDomain], 
                        trinity_required: bool) -> Optional[str]:
        """Recommend specific agent based on analysis."""
        if trinity_required:
            return "trinitas-coordinator"
        
        if len(domains) == 1:
            domain = domains[0]
            if domain == TrinityDomain.STRATEGY:
                return "springfield-strategist"
            elif domain == TrinityDomain.TECHNICAL:
                return "krukai-optimizer"
            elif domain == TrinityDomain.SECURITY:
                return "vector-auditor"
        
        # Default recommendation based on complexity
        if complexity == ComplexityLevel.COMPLEX:
            return "krukai-optimizer"
        elif complexity == ComplexityLevel.MODERATE:
            return "springfield-strategist"
        
        return None

def main():
    """Main function for command-line usage."""
    if len(sys.argv) < 2:
        print("Usage: trinity-classifier.py '<user_input>'")
        sys.exit(1)
    
    user_input = " ".join(sys.argv[1:])
    classifier = TrinityClassifier()
    result = classifier.analyze_input(user_input)
    
    # Output analysis result as JSON
    output = {
        "complexity": result.complexity.name,
        "domains": [d.name for d in result.domains],
        "confidence": result.confidence,
        "reasoning": result.reasoning,
        "trinity_required": result.trinity_required,
        "recommended_agent": result.recommended_agent
    }
    
    print(json.dumps(output, indent=2))

if __name__ == "__main__":
    main()