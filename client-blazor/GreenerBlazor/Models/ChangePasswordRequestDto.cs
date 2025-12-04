using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace GreenerBlazor.Models;

public class ChangePasswordRequestDto
{
    public string PasswordOld { get; set; } = null!;

    [Required]
    [MinLength(Util.PasswordLengthMin)]
    [MaxLength(Util.PasswordLengthMax)]
    [DataType(DataType.Password)]
    [RegularExpression(Util.PasswordRegex, ErrorMessage = Util.PasswordRegexError)]
    public string PasswordNew { get; set; } = null!;

    [Required]
    [MinLength(Util.PasswordLengthMin)]
    [MaxLength(Util.PasswordLengthMax)]
    [DataType(DataType.Password)]
    [RegularExpression(Util.PasswordRegex, ErrorMessage = Util.PasswordRegexError)]
    [Compare(nameof(PasswordNew), ErrorMessage = "Passwords do not match")]
    [JsonIgnore]
    public string ConfirmPasswordNew { get; set; } = null!;
}
