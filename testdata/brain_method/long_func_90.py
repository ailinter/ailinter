def long_function(data):
    """This function does way too much in 90+ lines."""
    result = []
    validated = []
    transformed = []
    aggregated = {}

    for item in data:
        if item is not None:
            if item.get("valid", False):
                validated.append(item)

    for item in validated:
        if item.get("type") == "A":
            transformed.append(item["value"] * 2)
        elif item.get("type") == "B":
            transformed.append(item["value"] * 3)
        else:
            transformed.append(item["value"])

    for val in transformed:
        if val > 100:
            if "high" not in aggregated:
                aggregated["high"] = []
            aggregated["high"].append(val)
        else:
            if "low" not in aggregated:
                aggregated["low"] = []
            aggregated["low"].append(val)

    for category in aggregated:
        total = sum(aggregated[category])
        count = len(aggregated[category])
        avg = total / count if count > 0 else 0
        result.append({
            "category": category,
            "total": total,
            "count": count,
            "average": avg,
        })

    overall_total = sum(r["total"] for r in result)
    overall_count = sum(r["count"] for r in result)

    summary = {
        "items": result,
        "grand_total": overall_total,
        "item_count": overall_count,
        "timestamp": "2024-01-01",
        "version": "1.0",
    }

    if overall_count > 0:
        summary["average"] = overall_total / overall_count

    return summary
