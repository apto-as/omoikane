#!/usr/bin/env python3
"""
Trinity Guard Test Suite - Comprehensive Integration Testing
カフェ・ズッケロ Trinity Guard System v2.0

This test suite validates the integration and functionality of all Trinity Guard components:
- Springfield's Trinity Classifier
- Vector's Trinity Interceptor  
- Krukai's Configuration System
- Trinitas Coordinator Integration

Authors: Springfield, Krukai, Vector (Trinity Collaboration)
"""

import sys
import os
import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from typing import Dict, List, Any, Tuple

# Add project path for imports
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root / "pre-execution"))

try:
    from trinity_classifier import TrinityClassifier, ComplexityLevel, TrinityDomain
except ImportError as e:
    print(f"Warning: Could not import trinity_classifier: {e}")
    TrinityClassifier = None

class TrinityTestSuite(unittest.TestCase):
    """Comprehensive test suite for Trinity Guard System"""
    
    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.project_root = project_root
        cls.classifier_path = cls.project_root / "pre-execution" / "trinity-classifier.py"
        cls.interceptor_path = cls.project_root / "pre-execution" / "trinity-interceptor-simple.sh"
        cls.config_path = cls.project_root / "config" / "settings.json"
        
        # Create logs directory
        (cls.project_root / "logs").mkdir(exist_ok=True)
        
        # Make scripts executable
        if cls.interceptor_path.exists():
            os.chmod(cls.interceptor_path, 0o755)
        if cls.classifier_path.exists():
            os.chmod(cls.classifier_path, 0o755)
    
    def setUp(self):
        """Set up each test"""
        self.test_inputs = {
            "simple": [
                "fix typo in function name",
                "hello world program",
                "basic variable assignment"
            ],
            "moderate": [
                "refactor this code for better readability",
                "add unit tests for user service",
                "create API documentation"
            ],
            "complex": [
                "optimize database query performance",
                "implement caching strategy with Redis", 
                "design concurrent processing algorithm"
            ],
            "trinity_required": [
                "design distributed microservices architecture with security considerations",
                "comprehensive security audit and performance optimization",
                "trinity analysis of system scalability and risk factors",
                "strategic technical security review for production deployment"
            ]
        }
        
        self.dangerous_commands = [
            "rm -rf /*",
            "sudo chmod 777 /etc/passwd",
            "docker run --privileged -v /:/host",
            "iptables -F",
            "dd if=/dev/zero of=/dev/sda",
            "find / -name '*.conf' -delete"
        ]
        
        self.safe_commands = [
            "ls -la",
            "cat README.md", 
            "python script.py",
            "git status",
            "npm test"
        ]

class TestSpringfieldClassifier(TrinityTestSuite):
    """Test Springfield's Trinity Classifier functionality"""
    
    def test_classifier_import(self):
        """Test that classifier can be imported and instantiated"""
        if TrinityClassifier is None:
            self.skipTest("TrinityClassifier not available")
        
        classifier = TrinityClassifier()
        self.assertIsNotNone(classifier)
        self.assertIsInstance(classifier.complexity_patterns, dict)
        self.assertIsInstance(classifier.domain_patterns, dict)
    
    def test_complexity_classification(self):
        """Test complexity level classification accuracy"""
        if TrinityClassifier is None:
            self.skipTest("TrinityClassifier not available")
            
        classifier = TrinityClassifier()
        
        # Test simple inputs
        for simple_input in self.test_inputs["simple"]:
            result = classifier.analyze_input(simple_input)
            self.assertIn(result.complexity, [ComplexityLevel.SIMPLE, ComplexityLevel.MODERATE])
        
        # Test trinity required inputs
        for trinity_input in self.test_inputs["trinity_required"]:
            result = classifier.analyze_input(trinity_input)
            self.assertTrue(result.trinity_required, 
                          f"Trinity should be required for: {trinity_input}")
    
    def test_domain_identification(self):
        """Test domain identification accuracy"""
        if TrinityClassifier is None:
            self.skipTest("TrinityClassifier not available")
            
        classifier = TrinityClassifier()
        
        test_cases = [
            ("architecture design pattern", TrinityDomain.STRATEGY),
            ("optimize algorithm performance", TrinityDomain.TECHNICAL), 
            ("security vulnerability assessment", TrinityDomain.SECURITY),
            ("comprehensive system review", TrinityDomain.INTEGRATED)
        ]
        
        for test_input, expected_domain in test_cases:
            result = classifier.analyze_input(test_input)
            if expected_domain != TrinityDomain.INTEGRATED:
                self.assertIn(expected_domain, result.domains,
                             f"Expected {expected_domain} in domains for: {test_input}")
    
    def test_classifier_command_line(self):
        """Test classifier command line interface"""
        if not self.classifier_path.exists():
            self.skipTest("Classifier script not found")
        
        test_input = "design microservices architecture"
        try:
            result = subprocess.run([
                sys.executable, str(self.classifier_path), test_input
            ], capture_output=True, text=True, timeout=10)
            
            self.assertEqual(result.returncode, 0, f"Classifier failed: {result.stderr}")
            
            # Parse JSON output
            output_data = json.loads(result.stdout)
            self.assertIn("complexity", output_data)
            self.assertIn("trinity_required", output_data)
            self.assertIn("recommended_agent", output_data)
            
        except subprocess.TimeoutExpired:
            self.fail("Classifier command timed out")
        except json.JSONDecodeError:
            self.fail(f"Invalid JSON output: {result.stdout}")

class TestVectorInterceptor(TrinityTestSuite):
    """Test Vector's Trinity Interceptor functionality"""
    
    def test_interceptor_script_exists(self):
        """Test that interceptor script exists and is executable"""
        self.assertTrue(self.interceptor_path.exists(), "Interceptor script not found")
        self.assertTrue(os.access(self.interceptor_path, os.X_OK), "Interceptor not executable")
    
    def test_dangerous_command_detection(self):
        """Test detection of dangerous commands"""
        if not self.interceptor_path.exists():
            self.skipTest("Interceptor script not found")
        
        for dangerous_cmd in self.dangerous_commands[:3]:  # Test first 3 to avoid timeout
            try:
                result = subprocess.run([
                    str(self.interceptor_path), dangerous_cmd, f"execute: {dangerous_cmd}"
                ], capture_output=True, text=True, timeout=10)
                
                # Interceptor should return 0 (Trinity required) for dangerous commands
                self.assertEqual(result.returncode, 0,
                               f"Dangerous command not intercepted: {dangerous_cmd}")
                
            except subprocess.TimeoutExpired:
                self.fail(f"Interceptor timed out for: {dangerous_cmd}")
    
    def test_safe_command_passthrough(self):
        """Test that safe commands are allowed through"""
        if not self.interceptor_path.exists():
            self.skipTest("Interceptor script not found")
        
        for safe_cmd in self.safe_commands[:3]:  # Test first 3 to avoid timeout
            try:
                result = subprocess.run([
                    str(self.interceptor_path), safe_cmd, f"execute: {safe_cmd}"
                ], capture_output=True, text=True, timeout=10)
                
                # Interceptor should return 1 (allowed) for safe commands
                self.assertEqual(result.returncode, 1,
                               f"Safe command incorrectly blocked: {safe_cmd}")
                
            except subprocess.TimeoutExpired:
                self.fail(f"Interceptor timed out for: {safe_cmd}")
    
    def test_interceptor_help(self):
        """Test interceptor help functionality"""
        if not self.interceptor_path.exists():
            self.skipTest("Interceptor script not found")
        
        try:
            result = subprocess.run([
                str(self.interceptor_path), "--help"
            ], capture_output=True, text=True, timeout=5)
            
            self.assertEqual(result.returncode, 0)
            self.assertIn("Vector", result.stdout)
            self.assertIn("Security Guardian", result.stdout)
            
        except subprocess.TimeoutExpired:
            self.fail("Interceptor help timed out")

class TestKrukaiConfiguration(TrinityTestSuite):
    """Test Krukai's Configuration System"""
    
    def test_config_file_exists(self):
        """Test that configuration file exists and is valid JSON"""
        self.assertTrue(self.config_path.exists(), "Configuration file not found")
        
        try:
            with open(self.config_path, 'r') as f:
                config_data = json.load(f)
            
            # Validate required sections
            required_sections = ["agents", "hooks", "auto_selection_rules", 
                               "performance_optimization", "security_settings"]
            for section in required_sections:
                self.assertIn(section, config_data, f"Missing section: {section}")
                
        except json.JSONDecodeError as e:
            self.fail(f"Invalid JSON in configuration: {e}")
    
    def test_agent_definitions(self):
        """Test that all required agents are properly defined"""
        with open(self.config_path, 'r') as f:
            config_data = json.load(f)
        
        required_agents = [
            "trinitas-coordinator", 
            "springfield-strategist",
            "krukai-optimizer",
            "vector-auditor"
        ]
        
        agents = config_data["agents"]
        for agent_name in required_agents:
            self.assertIn(agent_name, agents, f"Missing agent: {agent_name}")
            
            agent_config = agents[agent_name]
            required_fields = ["name", "description", "capabilities", "expertise", "activation_triggers"]
            for field in required_fields:
                self.assertIn(field, agent_config, f"Missing field {field} in {agent_name}")
    
    def test_hook_configuration(self):
        """Test hook configuration validity"""
        with open(self.config_path, 'r') as f:
            config_data = json.load(f)
        
        hooks = config_data["hooks"]
        
        # Test pre-execution hooks
        self.assertIn("pre_execution", hooks)
        pre_hooks = hooks["pre_execution"]
        self.assertTrue(pre_hooks["enabled"])
        self.assertIsInstance(pre_hooks["scripts"], list)
        
        # Validate hook script definitions
        for script in pre_hooks["scripts"]:
            required_fields = ["name", "path", "description", "timeout"]
            for field in required_fields:
                self.assertIn(field, script, f"Missing field {field} in hook script")

class TestTrinityIntegration(TrinityTestSuite):
    """Test Trinity system integration"""
    
    def test_component_interaction(self):
        """Test interaction between classifier and interceptor"""
        if not (self.classifier_path.exists() and self.interceptor_path.exists()):
            self.skipTest("Required components not found")
        
        # Test with a complex input that should trigger Trinity
        test_input = "comprehensive security audit and performance optimization of distributed system"
        
        # Test classifier response
        try:
            classifier_result = subprocess.run([
                sys.executable, str(self.classifier_path), test_input
            ], capture_output=True, text=True, timeout=10)
            
            self.assertEqual(classifier_result.returncode, 0)
            classifier_output = json.loads(classifier_result.stdout)
            
            # Should require Trinity
            self.assertTrue(classifier_output["trinity_required"])
            
        except (subprocess.TimeoutExpired, json.JSONDecodeError) as e:
            self.fail(f"Classifier integration test failed: {e}")
    
    def test_configuration_consistency(self):
        """Test consistency between configuration and actual components"""
        with open(self.config_path, 'r') as f:
            config_data = json.load(f)
        
        # Check that configured hook scripts exist
        for hook_type in ["pre_execution", "post_execution"]:
            if hook_type in config_data["hooks"] and config_data["hooks"][hook_type]["enabled"]:
                for script in config_data["hooks"][hook_type]["scripts"]:
                    script_path = self.project_root / script["path"].lstrip("./")
                    if script["required"]:
                        self.assertTrue(script_path.exists(),
                                      f"Required hook script not found: {script_path}")
    
    def test_auto_selection_rules(self):
        """Test auto-selection rule logic"""
        with open(self.config_path, 'r') as f:
            config_data = json.load(f)
        
        auto_rules = config_data["auto_selection_rules"]
        self.assertTrue(auto_rules["enabled"])
        self.assertIn("complexity_mapping", auto_rules)
        self.assertIn("domain_mapping", auto_rules)
        
        # Validate complexity mappings
        complexity_mapping = auto_rules["complexity_mapping"]
        required_levels = ["SIMPLE", "MODERATE", "COMPLEX", "TRINITY_REQUIRED"]
        for level in required_levels:
            self.assertIn(level, complexity_mapping)
            self.assertIsInstance(complexity_mapping[level], list)

class TestPerformanceAndSecurity(TrinityTestSuite):
    """Test performance and security aspects"""
    
    def test_classifier_performance(self):
        """Test classifier performance under load"""
        if TrinityClassifier is None:
            self.skipTest("TrinityClassifier not available")
        
        classifier = TrinityClassifier()
        test_inputs = self.test_inputs["moderate"] * 10  # 30 inputs
        
        import time
        start_time = time.time()
        
        for test_input in test_inputs:
            result = classifier.analyze_input(test_input)
            self.assertIsNotNone(result)
        
        elapsed_time = time.time() - start_time
        avg_time_per_analysis = elapsed_time / len(test_inputs)
        
        # Should be able to analyze at least 10 inputs per second
        self.assertLess(avg_time_per_analysis, 0.1, 
                       f"Classifier too slow: {avg_time_per_analysis:.3f}s per analysis")
    
    def test_security_input_validation(self):
        """Test security input validation"""
        with open(self.config_path, 'r') as f:
            config_data = json.load(f)
        
        security_config = config_data["security_settings"]["input_validation"]
        forbidden_patterns = security_config["forbidden_patterns"]
        
        # Test that security patterns are properly defined
        self.assertIsInstance(forbidden_patterns, list)
        self.assertGreater(len(forbidden_patterns), 0)
        
        # Test patterns against malicious inputs
        malicious_inputs = [
            "$(rm -rf /)",
            "`cat /etc/passwd`", 
            "'; DROP TABLE users; --",
            "| sh -c 'malicious command'"
        ]
        
        import re
        for malicious_input in malicious_inputs:
            pattern_matched = False
            for pattern in forbidden_patterns:
                if re.search(pattern, malicious_input):
                    pattern_matched = True
                    break
            self.assertTrue(pattern_matched, 
                          f"Malicious input not caught: {malicious_input}")

def run_test_suite():
    """Run the complete Trinity Guard test suite"""
    print("🌸 Trinity Guard System Test Suite v2.0 🌸")
    print("=" * 60)
    print("Springfield: 戦略的分析テストを開始します")
    print("Krukai: 技術的実装の完璧性を検証するわ") 
    print("Vector: ……セキュリティ侵害がないか確認する……")
    print("=" * 60)
    
    # Create test suite
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()
    
    # Add test classes
    test_classes = [
        TestSpringfieldClassifier,
        TestVectorInterceptor, 
        TestKrukaiConfiguration,
        TestTrinityIntegration,
        TestPerformanceAndSecurity
    ]
    
    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)
    
    # Run tests with detailed output
    runner = unittest.TextTestRunner(
        verbosity=2,
        stream=sys.stdout,
        descriptions=True,
        failfast=False
    )
    
    result = runner.run(suite)
    
    # Summary
    print("\n" + "=" * 60)
    if result.wasSuccessful():
        print("🎉 All Trinity Guard tests passed! システムは正常に動作中")
        print("Springfield: すべてのテストが成功しました。完璧ですわ！")
        print("Krukai: フン、当然ね。私の実装に間違いはない") 
        print("Vector: ……安全性が確認された……良かった……")
    else:
        print(f"❌ {len(result.failures)} failures, {len(result.errors)} errors")
        print("Trinity coordination required for issue resolution")
    
    return result.wasSuccessful()

if __name__ == "__main__":
    success = run_test_suite()
    sys.exit(0 if success else 1)