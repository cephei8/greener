using System.Text.Json.Serialization;

namespace GreenerBlazor.Models;

[JsonConverter(typeof(JsonStringEnumConverter<TestcaseStatus>))]
public enum TestcaseStatus
{
    Error,
    Fail,
    Pass,
    Skip,
}
