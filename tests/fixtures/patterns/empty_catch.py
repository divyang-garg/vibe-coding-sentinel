# Empty except block - should be detected
try:
    risky_operation()
except Exception:
    # Empty except block
    pass

# Another empty except
try:
    another_risky_op()
except:
    pass












