- `settings.py`で`TIME_ZONE = 'Asia/Tokyo'`に設定する必要がある
- `views.py`などでJSTの時刻を取得する方法
  ~~~python
  from django.utils import timezone

  datetime_now_jst = timezone.localtime(timezone.now(), timezone=timezone.get_default_timezone())
  ~~~