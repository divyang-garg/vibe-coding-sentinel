func validateUser(user User) error {
    if user.Email == "" {
        return fmt.Errorf("email required")
    }
    if len(user.Password) < 8 {
        return fmt.Errorf("password too short")
    }
    return nil
}
