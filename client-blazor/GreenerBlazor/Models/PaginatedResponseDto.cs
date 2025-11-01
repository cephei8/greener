namespace GreenerBlazor.Models;

public class PaginatedResponseDto<T>
{
    public required List<T> Items { get; set; }
    public required int Total { get; set; }
    public required int Limit { get; set; }
    public required int Offset { get; set; }
}
