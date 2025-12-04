using System.ComponentModel.DataAnnotations;

namespace GreenerBlazor.Models;

public class LoginRequestDto
{
    [Required]
    [MinLength(Util.UsernameLengthMin)]
    [MaxLength(Util.UsernameLengthMax)]
    [RegularExpression(Util.UsernameRegex, ErrorMessage = Util.UsernameRegexError)]
    public string Username { get; set; } = null!;

    [Required]
    [MinLength(Util.PasswordLengthMin)]
    [MaxLength(Util.PasswordLengthMax)]
    [DataType(DataType.Password)]
    [RegularExpression(Util.PasswordRegex, ErrorMessage = Util.PasswordRegexError)]
    public string Password { get; set; } = null!;
}
