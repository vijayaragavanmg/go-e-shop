-- 1) Add tsvector column (idempotent)
ALTER TABLE products
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 2) Create/replace function to update search_vector (includes category name at weight D)
CREATE OR REPLACE FUNCTION products_search_vector_update()
RETURNS trigger AS $$
DECLARE
    v_category_name text;
BEGIN
    -- Look up the category name for the NEW row
    SELECT c.name
      INTO v_category_name
      FROM categories c
     WHERE c.id = NEW.category_id;

    NEW.search_vector :=
          setweight(to_tsvector('english', coalesce(NEW.name, '')), 'A')
       || setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B')
       || setweight(to_tsvector('simple',  coalesce(NEW.sku, '')), 'C')     -- optional: 'simple' for SKU tokens
       || setweight(to_tsvector('english', coalesce(v_category_name, '')), 'D');

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3) (Re)create trigger to automatically update search_vector
-- Limit to relevant columns so we don't recompute on unrelated updates
DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;

CREATE TRIGGER products_search_vector_trigger
    BEFORE INSERT OR UPDATE OF name, description, sku, category_id
    ON products
    FOR EACH ROW
EXECUTE FUNCTION products_search_vector_update();

-- 4) Backfill existing rows WITHOUT referencing the function variable
--    Either call the trigger by updating a dependent column,
--    OR compute category name via subquery. We'll do the latter to be explicit.
WITH cat AS (
  SELECT p.id,
         setweight(to_tsvector('english', coalesce(p.name, '')), 'A')
      || setweight(to_tsvector('english', coalesce(p.description, '')), 'B')
      || setweight(to_tsvector('simple',  coalesce(p.sku, '')), 'C')
      || setweight(
           to_tsvector(
             'english',
             coalesce(
               (SELECT c.name FROM categories c WHERE c.id = p.category_id),
               ''
             )
           ),
           'D'
         ) AS sv
  FROM products p
)
UPDATE products p
SET search_vector = cat.sv
FROM cat
WHERE cat.id = p.id;

-- 5) Create a GIN index for fast FTS. Use IF NOT EXISTS if on PG 9.5+ with extension.
--    For large tables in production, prefer CONCURRENTLY (requires outside a transaction).
CREATE INDEX IF NOT EXISTS idx_products_search_vector
    ON products USING GIN (search_vector);

-- 6) Single, clear comment for documentation
COMMENT ON COLUMN products.search_vector IS
  'Full-text search vector (A=name, B=description, C=sku, D=category name)';
  