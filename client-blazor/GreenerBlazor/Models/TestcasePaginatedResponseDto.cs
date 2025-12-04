namespace GreenerBlazor.Models;

public class TestcasePaginatedResponseDto: PaginatedResponseDto<TestcaseDto>
{
    public required TestcaseStatus? AggregatedStatus { get; set; }
}
