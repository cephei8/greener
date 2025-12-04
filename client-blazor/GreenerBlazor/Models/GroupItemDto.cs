namespace GreenerBlazor.Models;

public class GroupItemDto
{
    public List<string?> Columns { get; set; } = [];
    public TestcaseStatus Status { get; set; }
}

public class GroupPaginatedResponseDto
{
    public required List<GroupItemDto> Items { get; set; }
    public required int Total { get; set; }
    public required int Limit { get; set; }
    public required int Offset { get; set; }
    public List<string>? Header { get; set; }
    public TestcaseStatus? AggregatedStatus { get; set; }
}
