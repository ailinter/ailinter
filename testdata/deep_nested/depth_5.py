def deep_nested(data):
    if data is not None:
        if data.is_active():
            for item in data.items:
                if item.is_valid():
                    if item.needs_processing():
                        process_item(item)
    return None
