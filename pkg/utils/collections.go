package utils

func ContainsString(
	l []string,
	elmt string,
) bool  {
  for _, e := range l{
    if e == elmt{
      return true
    }
  }

  return false;
}
