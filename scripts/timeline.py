
import matplotlib.pyplot as plt
from datetime import datetime
import random

events = [
    ("Nov 2015", "Apiary"),
    ("May 2018", "Apiary"),
    ("Sep 2018", "Apiary"),
    ("Sep 2019", "SessionM"),
    ("Mar 2021", "Virtual"),
    ("Jun 2021", "Virtual"),
    ("May 2022", "Productboard"),
    ("Jun 2022", "Productboard"),
    ("Nov 2022", "Sentinel One"),
    ("Feb 2023", "Pure Storage"),
    ("Apr 2023", "Outreach"),
    ("Jun 2023", "Betsys"),
    ("Sep 2023", "Heureka"),
    ("Feb 2024", "Pure Storage"),
    ("Apr 2024", "Betsys"),
    ("Jun 2024", "Livesport"),
    ("Oct 2024", "Sentinel One"),
    ("Nov 2024", "Pure Storage"),
    ("Mar 2025", "Pure Storage"),
    ("Jun 2025", "Pure Storage"),
    ("Oct 2025", "STRV"),
    ("Nov 2025", "Dataddo"),
]

def parse_date(d):
    fmts = ["%d.%m.%Y", "%b %Y", "%B %Y"]
    for fmt in fmts:
        try:
            return datetime.strptime(d, fmt)
        except:
            pass
    raise ValueError(f"Unknown date format: {d}")

dates = [parse_date(d) for d, _ in events]
labels = [label for _, label in events]

plt.figure(figsize=(50, 8))
plt.axhline(1, color='lightgray', linewidth=1)
plt.scatter(dates, [1]*len(dates), color='tab:blue', zorder=3)

for i, (date, label) in enumerate(zip(dates, labels)):
    offset = random.uniform(0.15, 0.5)
    offset = offset if i % 2 == 0 else -offset
    va = 'bottom' if i % 2 == 0 else 'top'
    y = 1 + offset
    plt.plot([date, date], [1, y], color='gray', linewidth=0.5, zorder=1)
    plt.text(date, y, label, rotation=0, ha='right', va=va, fontsize=9)

plt.yticks([])
plt.title("10-Year History of Go Prague Meetup")
# plt.tight_layout()
plt.show()
