namespace GreenerBlazor.Models;

public class TestcasePaginatedResponseDto
{
    public required List<TestcaseDto> Items { get; set; }
    public required int Total { get; set; }
    public required int Limit { get; set; }
    public required int Offset { get; set; }
    public required TestcaseStatus AggregatedStatus { get; set; }
}
