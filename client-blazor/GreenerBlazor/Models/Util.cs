namespace GreenerBlazor.Models;

public static class Util
{
    public const string UsernameRegex = "^[a-zA-Z0-9]+$";
    public const string UsernameRegexError = "Only letters and numbers are allowed.";

    public const string PasswordRegex = "^[a-zA-Z0-9@_.!\\-]*$";
    public const string PasswordRegexError = "Only letters, numbers, and @ _ . ! - are allowed.";

    public const int UsernameLengthMin = 1;
    public const int UsernameLengthMax = 32;

    public const int PasswordLengthMin = 6;
    public const int PasswordLengthMax = 32;
}
